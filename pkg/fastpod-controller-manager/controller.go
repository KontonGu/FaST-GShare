/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.

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

package fastpodcontrollermanager

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	fastpodscheme "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned/scheme"
	informers "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions/fastgshare.caps.in.tum/v1"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	// k8scontroller "k8s.io/kubernetes/pkg/controller"
)

const controllerAgentName = "fastpod-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a FaSTPod is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a FaSTPod fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by FaSTPod"
	// MessageResourceSynced is the message used for an Event fired when a FaSTPod
	// is synced successfully
	MessageResourceSynced = "FaSTPod synced successfully"

	faasKind       = "FaSTPod"
	FastGShareWarm = "fast-gshare/warmpool"

	FaSTPodLibraryDir  = "/fastpod/library"
	SchedulerIpFile    = FaSTPodLibraryDir + "/schedulerIP.txt"
	GPUClientPortStart = 58001
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc // can still create the key even the object is deleted;
)

type Controller struct {
	kubeClient    kubernetes.Interface
	fastpodClient clientset.Interface

	podsLister  corelisters.PodLister
	podsSynced  cache.InformerSynced
	podInformer coreinformers.PodInformer

	fastpodsLister listers.FaSTPodLister
	fastpodsSynced cache.InformerSynced

	nodesLister corelisters.NodeLister
	nodesSynced cache.InformerSynced

	// expectations *k8scontroller.UIDTrackingControllerExpectations

	pendingList    *list.List
	pendingListMux *sync.Mutex

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// containerdClient *containerd.Client
}

// NewController returns a new FaSTPod controller
func NewController(
	ctx context.Context,
	kubeclient kubernetes.Interface,
	fastpodclient clientset.Interface,
	nodeinformer coreinformers.NodeInformer,
	podinformer coreinformers.PodInformer,
	fastpodinformer informers.FaSTPodInformer) *Controller {

	// Create event broadcaster
	// Add fastpod-controller types to the default Kubernetes Scheme so Events can be
	// logged for fastpod-controller types.
	utilruntime.Must(fastpodscheme.AddToScheme(kubescheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClient:    kubeclient,
		fastpodClient: fastpodclient,

		podsLister:  podinformer.Lister(),
		podsSynced:  podinformer.Informer().HasSynced,
		podInformer: podinformer,

		fastpodsLister: fastpodinformer.Lister(),
		fastpodsSynced: fastpodinformer.Informer().HasSynced,

		nodesLister: nodeinformer.Lister(),
		nodesSynced: nodeinformer.Informer().HasSynced,

		// expectations: k8scontroller.NewUIDTrackingControllerExpectations(k8scontroller.NewControllerExpectations()),

		pendingList:    list.New(),
		pendingListMux: &sync.Mutex{},

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "FaSTPods"),
		recorder:  recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when FaSTPod resources change
	fastpodinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFaSTPod,
		UpdateFunc: func(old, new interface{}) {
			newFstp := new.(*fastpodv1.FaSTPod)
			oldFstp := old.(*fastpodv1.FaSTPod)
			klog.Infof("DEBUG: updating FaSTPod %s with replica %d ", newFstp.Name, *newFstp.Spec.Replicas)
			klog.Infof("DEBUG: queue length %d", controller.workqueue.Len())
			if newFstp.ResourceVersion == oldFstp.ResourceVersion {
				controller.enqueueFaSTPod(new)
				return
			}

			controller.enqueueFaSTPod(new)
		},
		DeleteFunc: controller.handleDeletedFaSTPod,
	})

	// Set up an event handler for when Pod resources change. This
	// handler will lookup the owner of the given Pod, and if it is
	// owned by a FaSTPod resource will enqueue that FaSTPod resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Pod resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	podinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*corev1.Pod)
			oldDepl := old.(*corev1.Pod)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				controller.handleObject(new)
				return
			}
			controller.handleObject(new)
		},
		//TODO release pod when scale down?
		DeleteFunc: controller.handleObject,
	})

	nodeinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controller.resourceChanged,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.

// Try out without stopCh, rather with context
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting FaSTPod controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.podsSynced, c.fastpodsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers", "count: ", workers)
	// Launch two workers to process FaSTPod resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}
	klog.Info("Started workers")
	<-ctx.Done()
	klog.Info("Shutting down workers")
	// pendingInsuranceTicker.Stop()
	// pendingInsuranceDone <- true
	return nil

}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {

}

func (c *Controller) enqueueFaSTPod(obj interface{}) {

}

func (c *Controller) handleDeletedFaSTPod(obj interface{}) {

}

func (c *Controller) handleObject(obj interface{}) {

}

func (c *Controller) resourceChanged(obj interface{}) {
	// push pending FaSTPods into workqueue
	c.pendingListMux.Lock()
	for p := c.pendingList.Front(); p != nil; p = p.Next() {
		c.workqueue.Add(p.Value)
	}
	c.pendingList.Init()
	c.pendingListMux.Unlock()
}
