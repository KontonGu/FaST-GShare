/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.
@Techinical University of Munich, CAPS Cloud Team

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	fastpodscheme "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned/scheme"
	informers "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

const (
	controllerAgentName   = "fastgshare-operator"
	faasKind              = "FaSTPod"
	functionPort          = 8080
	SuccessSynced         = "Synced"
	MessageResourceSynced = "FaSTPod synced successfully"
)

// Controller is the controller implementation for Function resources
type Controller struct {
	// kubeclient is a standard kubernetes clientset
	kubeclient kubernetes.Interface
	// fastclientset is a clientset for FaSTPod API group
	fastclientset  clientset.Interface
	fastpodsLister listers.FaSTPodLister

	// fastpodsSynced returns true if the pod store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	fastpodsSynced cache.InformerSynced

	nodelister corelisterv1.NodeLister

	nodesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	resolver *k8s.FunctionLookup

	// OpenFaaS function factory
	factory FunctionFactory
	//podQueue for handling events
	//podQueue
}

// NewController returns a new OpenFaaS controller
func NewController(
	kubeclientset kubernetes.Interface,
	fastpodclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	fastpodInformerFactory informers.SharedInformerFactory,
	factory FunctionFactory,
	resolver *k8s.FunctionLookup) *Controller {

	fastpodInformer := fastpodInformerFactory.Fastgshare().V1().FaSTPods()

	// Create event broadcaster
	// Add fastpod types to the default Kubernetes Scheme so Events can be
	// logged for faas-controller types.
	fastpodscheme.AddToScheme(scheme.Scheme)
	klog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclient:     kubeclientset,
		fastclientset:  fastpodclientset,
		fastpodsLister: fastpodInformer.Lister(),
		fastpodsSynced: fastpodInformer.Informer().HasSynced,
		workqueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Fastpods"),
		recorder:       recorder,
		resolver:       resolver,
		nodelister:     kubeInformerFactory.Core().V1().Nodes().Lister(),
		factory:        factory,
	}

	klog.Info("Setting up event handlers")

	//  Add Function (OpenFaaS CRD-entry) Informer
	// Set up an event handler for when Function resources change
	fastpodInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addFaSTPod,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFaSTPod(new)
		},
		DeleteFunc: controller.handleDeletedFaSTPod,
	})

	// Set up an event handler for when functions related resources like pods, deployments, replica sets
	// can't be materialized. This logs abnormal events like ImagePullBackOff, back-off restarting failed container,
	// failed to start container, oci runtime errors, etc
	kubeInformerFactory.Core().V1().Events().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				event := obj.(*corev1.Event)
				since := time.Since(event.LastTimestamp.Time)
				// log abnormal events occurred in the last minute
				if since.Seconds() < 61 && strings.Contains(event.Type, "Warning") {
					klog.V(3).Infof("Abnormal event detected on %s %s: %s", event.LastTimestamp, key, event.Message)
				}
			}
		},
	})

	kubeInformerFactory.Core().V1().Nodes().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//DeleteFunc: controller.resourceChanged,
	})

	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controller.deletePodInfo,
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (ctr *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer ctr.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ctr.fastpodsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Function resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(ctr.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the workqueue.
func (ctr *Controller) runWorker() {
	for ctr.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ctr *Controller) processNextWorkItem() bool {
	obj, shutdown := ctr.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer ctr.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			ctr.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := ctr.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		ctr.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (ctr *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the FaSTPod resource with this namespace/name
	fastpod, err := ctr.fastpodsLister.FaSTPods(namespace).Get(name)
	if err != nil {
		// The Function resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("fastpod '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	needUpdate := false

	if needUpdate {
		_, err := ctr.fastclientset.FastgshareV1().FaSTPods(namespace).Update(context.TODO(), fastpod, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Error updating fastpods %v/%v replica and selector...", namespace, name)
		}
	}

	// Get the deployment with the name specified in Function.spec
	//deployment, err := ctr.deploymentsLister.Deployments(fastpod.GetNamespace()).Get(deploymentName)
	// If the resource doesn't exist, we'll create it

	svcGetOptions := metav1.GetOptions{}
	funcname, found := fastpod.Labels["fast_function"]
	if !found {
		funcname = fastpod.ObjectMeta.Name
	}
	_, getSvcErr := ctr.kubeclient.CoreV1().Services(fastpod.Namespace).Get(context.TODO(), funcname, svcGetOptions)
	if errors.IsNotFound(getSvcErr) {
		klog.Infof("Creating ClusterIP service for '%s'", fastpod.Name)
		if _, err := ctr.kubeclient.CoreV1().Services(fastpod.Namespace).Create(context.TODO(), newService(fastpod), metav1.CreateOptions{}); err != nil {
			// If an error occurs during Service Create, we'll requeue the item
			if errors.IsAlreadyExists(err) {
				err = nil
				klog.V(2).Infof("ClusterIP service '%s' already exists. Skipping creation.", fastpod.Name)
			} else {
				return err
			}
		}
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return fmt.Errorf("transient error: %v", err)
	}

	existingService, err := ctr.kubeclient.CoreV1().Services(fastpod.Namespace).Get(context.TODO(), funcname, metav1.GetOptions{})
	if err != nil {
		return err
	}

	existingService.Annotations = makeAnnotations(fastpod)
	_, err = ctr.kubeclient.CoreV1().Services(fastpod.Namespace).Update(context.TODO(), existingService, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Updating service for '%s' failed: %v", funcname, err)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	ctr.recorder.Event(fastpod, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (ctr *Controller) addFaSTPod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		//handle error
		runtime.HandleError(err)
	}

	fastpod, err := ctr.fastclientset.FastgshareV1().FaSTPods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Error getting FaSTPod= %v", name)
	}

	//create the map for this fastpod/function
	ctr.resolver.AddFunc(fastpod.Name)

	if len(fastpod.Spec.PodSpec.InitContainers) > 0 {
		klog.Infof("Starting to create init container for the FaSTPod %s/%s", fastpod.Namespace, fastpod.Name)

		_, err := ctr.kubeclient.AppsV1().DaemonSets(namespace).Create(context.TODO(), newDaemonset(fastpod), metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Error %v starting init container for the fastpod %v/%v", err, namespace, name)
			runtime.HandleError(err)

		}
	}
	ctr.workqueue.AddRateLimited(key)
}

func newDaemonset(fastpod *fastpodv1.FaSTPod) *appsv1.DaemonSet {
	namePrefix := fastpod.Name + "-init"

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namePrefix,
			Namespace: fastpod.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"init": fastpod.Name}},
			//TTLSecondsAfterFinished: int32p(100),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namePrefix,
					Namespace: fastpod.Namespace,
					Labels:    map[string]string{"init": fastpod.Name},
				},
				Spec: corev1.PodSpec{

					//NodeName:              node,
					Containers:    fastpod.Spec.PodSpec.InitContainers,
					Volumes:       fastpod.Spec.PodSpec.Volumes,
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
}

func (ctr *Controller) enqueueFaSTPod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	ctr.workqueue.Add(key)
}

func (ctr *Controller) handleDeletedFaSTPod(obj interface{}) {
	fastpod, ok := obj.(*fastpodv1.FaSTPod)

	if !ok {
		runtime.HandleError(fmt.Errorf("handleDeletedFaSTPod: cannot parse object"))
		return
	}

	namespace := fastpod.Namespace
	name := fastpod.Name

	klog.Infof("Fastpod %v/%v deleted...", namespace, name)
	ctr.resolver.DeleteFunction(name)
}

func (ctr *Controller) deletePodInfo(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		if ownerRef := metav1.GetControllerOf(pod); ownerRef != nil {
			if ownerRef.Kind == "FaSTPod" {
				ctr.resolver.DeletePodInfo(ownerRef.Name, pod.Name)
			}
		}
	}
}
