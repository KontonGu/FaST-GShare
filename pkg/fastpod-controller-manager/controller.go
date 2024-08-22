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

package fastpodcontrollermanager

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	fastpodscheme "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned/scheme"
	informers "github.com/KontonGu/FaST-GShare/pkg/client/informers/externalversions/fastgshare.caps.in.tum/v1"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	k8scontroller "k8s.io/kubernetes/pkg/controller"
	"k8s.io/utils/integer"
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

	fastpodKind    = "FaSTPod"
	FastGShareWarm = "fast-gshare/warmpool"

	FaSTPodLibraryDir  = "/fastpod/library"
	SchedulerIpFile    = FaSTPodLibraryDir + "/schedulerIP.txt"
	GPUClientPortStart = 56001
	GPUSchedPortStart  = 52001

	ErrValueError             = "ErrValueError"
	SlowStartInitialBatchSize = 1
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

	expectations *k8scontroller.UIDTrackingControllerExpectations

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

	eventBroadcaster := record.NewBroadcaster()
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

		expectations: k8scontroller.NewUIDTrackingControllerExpectations(k8scontroller.NewControllerExpectations()),

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
// as syncing informer caches and starting workers. It will block until context
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.

func (ctr *Controller) Run(stopCh <-chan struct{}, workers int) error {
	defer utilruntime.HandleCrash()
	defer ctr.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting FaSTPod controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ctr.podsSynced, ctr.fastpodsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	ctr.gpuNodeInit()

	go ctr.startConfigManager(stopCh, ctr.kubeClient)
	klog.Infof("Starting workers, Numuber of workers = %d.", workers)
	// Launch two workers to process FaSTPod resources
	for i := 0; i < workers; i++ {
		go wait.Until(ctr.runWorker, time.Second, stopCh)
	}
	klog.Info("Workers Started")
	<-stopCh
	klog.Info("Shutting down workers")
	// pendingInsuranceTicker.Stop()
	// pendingInsuranceDone <- true
	return nil

}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (ctr *Controller) runWorker() {
	for ctr.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ctr *Controller) processNextWorkItem() bool {
	objRef, shutdown := ctr.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer ctr.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			ctr.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("The object in the workqueue is not valid. obj=%#v", obj))
			return nil
		}
		if err := ctr.syncHandler(key); err != nil {
			if err.Error() == "Waiting4Dummy" {
				ctr.workqueue.Add(key)
				return fmt.Errorf("TESTING: need to wait for dummy pod '#{key}', requeueing")
			}

			ctr.workqueue.AddRateLimited(key)
			return fmt.Errorf("Error while syncing the object = %s: %s. The object is re-queued.", key, err.Error())
		}

		ctr.workqueue.Forget(obj)
		klog.Infof("Successfully sync the object = %s.", key)
		return nil
	}(objRef)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the FaSTPod resource
// with the current status of the resource.
func (ctr *Controller) syncHandler(key string) error {
	startTime := time.Now()
	klog.V(2).Infof("Starting to sync FaSTPod %q (%v)", key, time.Since(startTime))
	defer func() {
		klog.V(4).Infof("Finished syncing FaSTPod %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Error invalid resource key = %s.", key))
		return nil
	}

	fastpod, err := ctr.fastpodsLister.FaSTPods(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("Error FaSTPod '%s' is no longer existed when syncing it.", key))
			return nil
		}
		return err
	}

	if fastpod.Spec.Replicas == nil {
		klog.Infof("Waiting FaSTPod %v/%v replicas to be updated ...", namespace, name)
		return nil
	}

	fastpodCopy := fastpod.DeepCopy()

	if fastpodCopy.Spec.Selector == nil {
		fastpodCopy.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":        fastpodCopy.Name,
				"controller": fastpodCopy.Name,
			},
		}
	}

	if fastpodCopy.Status.BoundDeviceIDs == nil {
		boundIds := make(map[string]string)
		fastpodCopy.Status.BoundDeviceIDs = &boundIds
	}

	if fastpodCopy.Status.Pod2Node == nil {
		pod2Node := make(map[string]string)
		fastpodCopy.Status.Pod2Node = &pod2Node
	}

	if fastpodCopy.Status.GPUClientPort == nil {
		gpuclientPort := make(map[string]int)
		fastpodCopy.Status.GPUClientPort = &gpuclientPort
	}

	if fastpodCopy.Status.Usage == nil {
		usages := make(map[string]fastpodv1.FaSTPodUsage)
		fastpodCopy.Status.Usage = &usages
	}

	syncFaSTPod := true
	selector, err := metav1.LabelSelectorAsSelector(fastpodCopy.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting pod selector to selector for the fastpod %v/%v: %v", namespace, name, err))
	}
	if selector == nil {
		klog.Errorf("Error the selector of fastpod %s/%s is till nil...", namespace, name)
		return nil
	}

	// list pods of a FaSTPod
	allPods, err := ctr.podsLister.Pods(namespace).List(selector)
	if err != nil {
		klog.Errorf("Error cannot get pods of the FaSTPod = %s.", key)
		return err
	}

	// Ignore inactive pods
	filteredPods := filterInactivePods(allPods)
	klog.Infof("The FaSTPod=%s/%s now has %s pods.", fastpodCopy.Namespace, fastpodCopy.Name, len(filteredPods))

	// reconcile the replicas of the fastpod
	var manageReplicasErr error
	if syncFaSTPod {
		manageReplicasErr = ctr.reconcileReplicas(context.TODO(), filteredPods, fastpodCopy, key)
	}
	newStatus := getFaSTPodReplicaStatus(fastpodCopy, filteredPods, manageReplicasErr)

	fastpodCopy.Status.AvailableReplicas = newStatus.AvailableReplicas
	fastpodCopy.Status.ReadyReplicas = newStatus.ReadyReplicas
	fastpodCopy.Status.Replicas = newStatus.Replicas

	updatedFastpod, err := ctr.fastpodClient.FastgshareV1().FaSTPods(fastpodCopy.Namespace).Update(context.TODO(), fastpodCopy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	if manageReplicasErr != nil && updatedFastpod.Status.ReadyReplicas == *(updatedFastpod.Spec.Replicas) &&
		updatedFastpod.Status.AvailableReplicas != *(updatedFastpod.Spec.Replicas) {
		ctr.enqueueFaSTPod(updatedFastpod)
	}
	return manageReplicasErr
}

// reconcile the spec.Replicas and existed replcias
func (ctr *Controller) reconcileReplicas(ctx context.Context, existedPods []*corev1.Pod, fastpod *fastpodv1.FaSTPod, key string) error {

	diff := len(existedPods) - int(*(fastpod.Spec.Replicas))
	klog.Infof("Current FaSTPod = %s has %d replicas with specification spec = %d, diff = %d.", key, len(existedPods), int(*(fastpod.Spec.Replicas)), diff)

	fstpKey, err := KeyFunc(fastpod)
	if err != nil {
		utilruntime.HandleError((fmt.Errorf("Error failed to get key of FaSTPod = %v %#v: %v.", fastpod.Kind, fastpod, err)))
		return nil
	}

	fastpodCopy := fastpod.DeepCopy()

	fstp2PodsMtx.Lock()
	defer fstp2PodsMtx.Unlock()

	// To create new pods if replicas is not enough
	if diff < 0 {
		// the number of pods to create
		diff *= -1
		ctr.expectations.ExpectCreations(fstpKey, diff)
		klog.V(2).Infof("Not enough replicas for the FaSTPod ... \n need %d replicas, try to create %d replicas", *fastpodCopy.Spec.Replicas, diff)
		successedNum, err := slowStartbatch(diff, k8scontroller.SlowStartInitialBatchSize, func() (*corev1.Pod, error) {
			isValidFastpod := false
			quotaReq := 0.0
			quotaLimit := 0.0
			smPartition := int64(100)
			gpuMem := int64(0)

			gpuDevUUID := ""
			gpuClientPort := 0

			// check the validity of fastpod resource configuration and get the resource configuration for a pod of FaSTPod
			if fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaRequest] != "" ||
				fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaLimit] != "" ||
				fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUSMPartition] != "" {
				var err error
				objName := fastpod.ObjectMeta.Name
				objNamesapce := fastpod.ObjectMeta.Namespace
				tmpQLStr := fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaLimit]
				quotaLimit, err = strconv.ParseFloat(tmpQLStr, 64)
				if err != nil || quotaLimit > 1.0 || quotaLimit < 0.0 {
					utilruntime.HandleError(fmt.Errorf("Error The FaSTPod = %s/%s has invalid quota limitation value %s.", objNamesapce, objName, tmpQLStr))
					return nil, err
				}

				tmpQRStr := fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaRequest]
				quotaReq, err = strconv.ParseFloat(tmpQRStr, 64)
				if err != nil || quotaReq > 1.0 || quotaReq < 0.0 {
					utilruntime.HandleError(fmt.Errorf("Error The FaSTPod = %s/%s has invalid quota request value %s.", objNamesapce, objName, quotaReq))
					return nil, err
				}

				tmpSMPaStr := fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUSMPartition]
				smPartition, err = strconv.ParseInt(tmpSMPaStr, 10, 64)
				if err != nil || smPartition < 0 || smPartition > 100 {
					utilruntime.HandleError(fmt.Errorf("Error The FaSTPod = %s/%s has invalid SM partition value %s.", objNamesapce, objName, tmpSMPaStr))
					smPartition = int64(100)
				}

				tmpMemStr := fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUMemory]
				gpuMem, err = strconv.ParseInt(tmpMemStr, 10, 64)
				if err != nil || gpuMem < 0 {
					utilruntime.HandleError(fmt.Errorf("Error The FaSTPod = %s/%s has invalid memory value %s.", objNamesapce, objName, gpuMem))
				}
				isValidFastpod = true
			}

			// get the node and gpu id (vGPU ID) the pod should be scheduled to based on the scheduling algorithm
			var schedNode, schedvGPUID string
			schedNode, schedvGPUID = ctr.schedule(fastpod, quotaReq, quotaLimit, smPartition, gpuMem, isValidFastpod, key)
			if schedNode == "" {
				return nil, errors.New("NoSchedNodeAvailable")
			}
			klog.Infof("The pod of FaSTPod = %s is scheduled to the node = %s with GPUID = %s", key, schedNode, schedvGPUID)

			// generate the pod key for the new pod of FaSTPod
			var subpodName string
			var subpodKey string
			if isValidFastpod {
				var errCode int
				fstpName := fastpodCopy.Name
				if fstp2Pods[fstpName] == nil {
					fstp2Pods[fstpName] = list.New()
				}
				newPodName := fstpName + "-" + RandStr(5)
				subpodName = newPodName
				subpodKey = fmt.Sprintf("%s/%s", fastpodCopy.ObjectMeta.Namespace, subpodName)

				// get the gpu device uuid and update the pod resource configuration in configurator
				gpuDevUUID, errCode = ctr.getGPUDevUUIDAndUpdateConfig(schedNode, schedvGPUID, quotaReq, quotaLimit, smPartition, gpuMem, subpodKey, &gpuClientPort)
				klog.Infof("The pod = %s of FaSTPod %s with vGPUID = %s is bound to device UUID=%s with GPUClientPort=%s.", subpodKey, key, schedvGPUID, gpuDevUUID, gpuClientPort)

				// errCode 0: no error
				// errCode 1: node with nodeName is not initialized
				// errCode 2: vGPUID is not initialized or no DummyPod created;
				// errCode 3: resource exceed;
				// errCode 4: GPU is out of memory
				// errCode 5: No enough gpu client ports
				switch errCode {
				case 0:
					klog.Infof("The pod is successfully bound.")
					fstp2Pods[fstpName].PushBack(newPodName)
				case 1:
					return nil, errors.New("NodeNotInitialized")
				case 2:
					return nil, errors.New("Waiting4Dummy")
				case 3:
					err := fmt.Errorf("Compute Resource exceed!")
					utilruntime.HandleError(err)
					ctr.recorder.Event(fastpod, corev1.EventTypeWarning, ErrValueError, "Compute Resource exceed")
					return nil, err
				case 4:
					err := fmt.Errorf("Out of memory!")
					utilruntime.HandleError(err)
					ctr.recorder.Event(fastpod, corev1.EventTypeWarning, ErrValueError, "Out of memory")
					return nil, err
				case 5:
					err := fmt.Errorf("GPU Clients Port is full!")
					utilruntime.HandleError(err)
					ctr.recorder.Event(fastpod, corev1.EventTypeWarning, ErrValueError, "GPU Clients Port is not enough")
					return nil, err
				default:
					err := fmt.Errorf("Unknown Error")
					utilruntime.HandleError(err)
					ctr.recorder.Event(fastpod, corev1.EventTypeWarning, ErrValueError, "Unknown Error")
					return nil, err
				}
			}

			// Create the new pod for the fastpod
			if node, ok := nodesInfo[schedNode]; ok {
				klog.Infof("Starting to create a new pod=%s of the fastpod=%s", subpodName, key)
				newpod, err := ctr.kubeClient.CoreV1().Pods(fastpodCopy.Namespace).Create(context.TODO(), ctr.newPod(fastpod, false, node.DaemonIP, gpuClientPort, gpuDevUUID, schedNode, schedvGPUID, subpodName), metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("Error when creating pod=%s for the FaSTPod=%s/%s.", subpodName, fastpod.Namespace, fastpod.Name)
					if apierrors.HasStatusCause(err, corev1.NamespaceTerminatingCause) {
						return nil, nil
					}
					return nil, err
				}
				(*fastpod.Status.BoundDeviceIDs)[newpod.Name] = schedvGPUID
				(*fastpod.Status.GPUClientPort)[newpod.Name] = gpuClientPort
				return newpod, err
			}

			return nil, nil

		})

		// to do
		if skippedPodsNum := diff - successedNum; skippedPodsNum > 0 {
			klog.V(2).Infof("The controller does not create enough pod for FaSTPod=%s, created Number:%d, falied number:%d.", key, successedNum, diff-successedNum)
			for i := 0; i < skippedPodsNum; i++ {

				//Decrement the expected number of creates because the informer won't observe this pod
				ctr.expectations.CreationObserved(fstpKey)
			}
		}
		return err
	} else if diff > 0 { // Too many Replicas, to delete pod to reconcile the spec
		klog.V(2).Infof("Too many replicas for the FaSTPod ... \n need %d replicas, try to delete %d replicas", *fastpodCopy.Spec.Replicas, diff)
		podsToDelete := ctr.getPodsToDelete(existedPods, diff)

		if podsToDelete == nil {
			klog.V(2).Infof("The number of pods=%d to delete exceeds the existed pods=%d.", diff, len(existedPods))
		}

		ctr.expectations.ExpectDeletions(key, getPodKeys(podsToDelete))

		errCh := make(chan error, diff)

		var wg sync.WaitGroup
		wg.Add(diff)
		for _, pod := range podsToDelete {
			go func(targetPod *corev1.Pod) {
				podCopy := targetPod.DeepCopy()
				defer func() {
					wg.Done()
					ctr.removePodFromList(fastpodCopy, podCopy)
				}()
				if err := ctr.kubeClient.CoreV1().Pods(targetPod.Namespace).Delete(ctx, targetPod.Name, metav1.DeleteOptions{}); err != nil {
					podKey := k8scontroller.PodKey(targetPod)
					ctr.expectations.DeletionObserved(key, podKey)
					if !apierrors.IsNotFound(err) {
						klog.V(2).Infof("Failed to delete pod=%s of the FaSTPod=%s.", podKey, key)
						errCh <- err
					}
				}
			}(pod)
		}
		wg.Wait()
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		default:
		}

	}

	return nil
}

// slowStartBatch tries to call the provided function a total of 'count' times,
// starting slow to check for errors, then speeding up if calls succeed.
//
// It groups the calls into batches, starting with a group of initialBatchSize.
// Within each batch, it may call the function multiple times concurrently.
//
// If a whole batch succeeds, the next batch may get exponentially larger.
// If there are any failures in a batch, all remaining batches are skipped
// after waiting for the current batch to complete.
//
// It returns the number of successful calls to the function.
func slowStartbatch(count int, initailBatchSize int, fn func() (*corev1.Pod, error)) (int, error) {
	remaining := count
	successes := 0
	need2wait := 0
	for batchSize := integer.IntMin(remaining, initailBatchSize); batchSize > 0; batchSize = integer.IntMin(2*batchSize, remaining) {
		errCh := make(chan error, batchSize)
		var wg sync.WaitGroup
		wg.Add(batchSize)
		for i := 0; i < batchSize; i++ {
			func() {
				defer wg.Done()
				if pod, err := fn(); err != nil {
					if err.Error() != "Waiting4Dummy" {
						//continue
						errCh <- err
					}
					if pod == nil && err.Error() == "Waiting4Dummy" {
						need2wait++
					}

				}
			}()
		}

		wg.Wait()
		curSuccesses := batchSize - len(errCh) - need2wait
		successes += curSuccesses
		if len(errCh) > 0 {
			return successes, <-errCh
		}
		remaining -= batchSize
	}
	if need2wait > 0 {
		return successes, errors.New("Waiting4Dmmy")
	}
	return successes, nil
}

// enqueue the object to the controller's workqueue
func (ctr *Controller) enqueueFaSTPod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	ctr.workqueue.Add(key)
}

func (ctr *Controller) handleDeletedFaSTPod(obj interface{}) {

}

func (ctr *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleError(fmt.Errorf("Error decoding object, invalid type, type = %T", obj))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleError(fmt.Errorf("Error decoding object tombstone, invalid type, type = %T", tombstone.Obj))
			return
		}
		klog.V(4).Info("Recovered deleted object", "resourceName", object.GetName())
	}

	// TODO logic to handle dummyPod if the status is "PodFailed"
	// dummyPod handling

	// enqueue FaSTPod key if it is a pod of a FaSTPod
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != fastpodKind {
			return
		}
		klog.V(4).Info("Processing object ", "object = ", klog.KObj(object))
		fastpod, err := ctr.fastpodsLister.FaSTPods(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("Ignoring orphaned object '%s' of FaSTPod '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}
		klog.Infof("The pod=%s of the FaSTPod = %s is to be processed ...", object.GetName(), ownerRef.Name)
		ctr.enqueueFaSTPod(fastpod)
		return
	}

}

func (ctr *Controller) resourceChanged(obj interface{}) {
	// push pending FaSTPods into workqueue
	ctr.pendingListMux.Lock()
	for p := ctr.pendingList.Front(); p != nil; p = p.Next() {
		ctr.workqueue.Add(p.Value)
	}
	ctr.pendingList.Init()
	ctr.pendingListMux.Unlock()
}

// newPod create a new pod specification based on the given information for the FaSTPod
func (ctr *Controller) newPod(fastpod *fastpodv1.FaSTPod, isWarm bool, schedIP string, gpuClientPort int, boundDevUUID, schedNode, schedvGPUID, podName string) *corev1.Pod {
	specCopy := fastpod.Spec.PodSpec.DeepCopy()
	specCopy.NodeName = schedNode

	labelCopy := makeLabels(fastpod)
	annotationCopy := make(map[string]string, len(fastpod.ObjectMeta.Annotations)+5)
	for key, val := range fastpod.ObjectMeta.Annotations {
		annotationCopy[key] = val
	}

	// TODO pre-warm setting for cold start issues, currently TODO...
	if isWarm {
		annotationCopy[FastGShareWarm] = "true"
	} else {
		annotationCopy[FastGShareWarm] = "false"
	}

	for i := range specCopy.Containers {
		ctn := &specCopy.Containers[i]
		ctn.Env = append(ctn.Env,
			corev1.EnvVar{
				Name:  "NVIDIA_VISIBLE_DEVICES",
				Value: boundDevUUID,
			},
			corev1.EnvVar{
				Name:  "NVIDIA_DRIVER_CAPABILITIES",
				Value: "compute,utility",
			},
			corev1.EnvVar{
				Name:  "LD_PRELOAD",
				Value: FaSTPodLibraryDir + "/libfast.so.1",
			},
			corev1.EnvVar{
				// the scheduler IP is not necessary since the hooked containers get it from /fastpod/library/GPUClientsIP.txt
				Name:  "SCHEDULER_IP",
				Value: schedIP,
			},
			corev1.EnvVar{
				Name:  "GPU_CLIENT_PORT",
				Value: fmt.Sprintf("%d", gpuClientPort),
			},
			corev1.EnvVar{
				Name:  "POD_NAME",
				Value: fmt.Sprintf("%s/%s", fastpod.ObjectMeta.Namespace, podName),
			},
		)
		ctn.VolumeMounts = append(ctn.VolumeMounts,
			corev1.VolumeMount{
				Name:      "fastpod-lib",
				MountPath: FaSTPodLibraryDir,
			},
			corev1.VolumeMount{
				Name:      "nvidia-mps",
				MountPath: "/tmp/nvidia-mps",
			},
		)
		ctn.ImagePullPolicy = fastpod.Spec.PodSpec.Containers[0].ImagePullPolicy
	}

	specCopy.Volumes = append(specCopy.Volumes,
		corev1.Volume{
			Name: "fastpod-lib",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: FaSTPodLibraryDir,
				},
			},
		},
		corev1.Volume{
			Name: "nvidia-mps",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/tmp/nvidia-mps",
				},
			},
		},
	)
	annotationCopy[fastpodv1.FaSTGShareGPUQuotaRequest] = fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaRequest]
	annotationCopy[fastpodv1.FaSTGShareGPUQuotaLimit] = fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUQuotaLimit]
	annotationCopy[fastpodv1.FaSTGShareGPUMemory] = fastpod.ObjectMeta.Annotations[fastpodv1.FaSTGShareGPUMemory]
	annotationCopy[fastpodv1.FaSTGShareVGPUID] = schedvGPUID

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: fastpod.ObjectMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(fastpod, schema.GroupVersionKind{
					Group:   fastpodv1.SchemeGroupVersion.Group,
					Version: fastpodv1.SchemeGroupVersion.Version,
					Kind:    fastpodKind,
				}),
			},
			Annotations: annotationCopy,
			Labels:      labelCopy,
		},
		Spec: corev1.PodSpec{
			NodeName:   schedNode,
			Containers: specCopy.Containers,
			Volumes:    specCopy.Volumes,
			HostIPC:    true,
			//InitContainers: []corev1.Container{},
		},
	}
}
