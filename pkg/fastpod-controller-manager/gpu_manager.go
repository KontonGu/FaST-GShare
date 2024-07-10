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

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	"github.com/KontonGu/FaST-GShare/pkg/lib/bitmap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

type PodReq struct {
	Key           string
	QtRequest     float64
	QtLimit       float64
	SMPartition   int64
	Memory        int64
	GPUClientPort int
}

type GPUDevInfo struct {
	GPUType string
	UUID    string
	Mem     int64
	// Usage of GPU resource, SM * QtRequest
	Usage   float64
	PodList *list.List
}

type NodeStatusInfo struct {
	// The available GPU device number
	GPUNum int32
	// The IP to the node Configurator and Scheduler Daemon
	DaemonIP string
	// The mapping from Physical GPU device UUID to fast-scheduler port
	UUID2SchedPort map[string]string
	// The mapping from vGPU ID (Client GPU) to physical GPU device
	vGPUID2GPU map[string]*GPUDevInfo
	// The port allocator for gpu clients, fast schedulers, and the configurator
	DaemonPortAlloc *bitmap.Bitmap
}

var (
	nodesInfo    map[string]*NodeStatusInfo
	nodesInfoMtx sync.Mutex
)

var Quantity1 = resource.MustParse("1")

func (ctr *Controller) gpuNodeInit() error {
	var nodes []*corev1.Node
	var err error

	if nodes, err = ctr.nodesLister.List(labels.Set{"gpu": "present"}.AsSelector()); err != nil {
		errr := fmt.Errorf("error when list nodes: #{err}")
		klog.Error(errr)
		return errr
	}

	type dummyPodInfo struct {
		nodename string
		vgpuID   string
	}

	var dummyPodInfoList []dummyPodInfo
	for _, node := range nodes {
		infoItem := dummyPodInfo{
			nodename: node.Name,
			vgpuID:   fastpodv1.GenerateGPUID(6),
		}
		dummyPodInfoList = append(dummyPodInfoList, infoItem)
		if nodeItem, ok := nodesInfo[node.Name]; !ok {
			pBm := bitmap.NewBitmap(256)
			pBm.Set(0)
			nodeItem = &NodeStatusInfo{
				vGPUID2GPU: make(map[string]*GPUDevInfo),
				// DaemonPortAlloc: pBm,
			}
			nodeItem.vGPUID2GPU[infoItem.vgpuID] = &GPUDevInfo{
				GPUType: "",
				UUID:    "",
				Mem:     0,
				Usage:   0.0,
				PodList: list.New(),
			}
			nodesInfo[node.Name] = nodeItem
		}
	}

	for _, item := range dummyPodInfoList {
		go ctr.createDummyPod(item.nodename, item.vgpuID)
	}
	return nil
}

func (ctr *Controller) createDummyPod(nodeName, vgpuID string) error {
	dummypodName := fmt.Sprintf("%s-%s-%s", fastpodv1.FaSTGShareDummpyPodName, nodeName, vgpuID)

	// function to create the dummy pod
	createFunc := func() error {
		dummypod, err := ctr.kubeClient.CoreV1().Pods("kube-system").Create(context.TODO(), &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dummypodName,
				Namespace: "kube-system",
				Labels: map[string]string{
					fastpodv1.FaSTGShareRole:     "dummyPod",
					fastpodv1.FaSTGShareNodeName: nodeName,
					fastpodv1.FaSTGShareVGPUID:   vgpuID,
				},
			},
			Spec: corev1.PodSpec{
				NodeName: nodeName,
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "dummy-gpu-acquire",
						Image: "kontonpuku666/dummycontainer:release",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{fastpodv1.OriginalNvidiaResourceName: Quantity1},
							Limits:   corev1.ResourceList{fastpodv1.OriginalNvidiaResourceName: Quantity1},
						},
					},
				},
				TerminationGracePeriodSeconds: new(int64),
				RestartPolicy:                 corev1.RestartPolicyNever,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			_, ok := ctr.kubeClient.CoreV1().Pods("kube-system").Get(context.TODO(), dummypodName, metav1.GetOptions{})
			if ok != nil {
				klog.Errorf("Create DummyPod failed: dummyPodName = %s\n, podSpec = \n %-v, \n err = '%s'.", dummypodName, dummypod, err)
				return err
			}
		}
		return nil
	}
	klog.Infof("DummyPod = %s is created successfully.", dummypodName)

	if existedpod, err := ctr.kubeClient.CoreV1().Pods("kube-system").Get(context.TODO(), dummypodName, metav1.GetOptions{}); err != nil {
		// The dummypod is not existed currently
		if errors.IsNotFound(err) {
			tocreate := true
			nodesInfoMtx.Lock()
			_, tocreate = nodesInfo[nodeName].vGPUID2GPU[vgpuID]
			nodesInfoMtx.Unlock()
			if tocreate {
				createFunc()
			}
		} else {
			// other reason except the NotFound
			klog.Errorf("Error when trying to get the DummyPod = %s, not the NotFound Error.\n", dummypodName)
			return err
		}
	} else {
		if existedpod.ObjectMeta.DeletionTimestamp != nil {
			// TODO: If Dummy Pod had been deleted, re-create it later
			klog.Warningf("Unhandled: Dummy Pod %s is deleting! re-create it later!", dummypodName)
		}
		if existedpod.Status.Phase == corev1.PodRunning || existedpod.Status.Phase == corev1.PodFailed {
			// TODO logic if dummy Pod is running or PodFailed
		}
	}
	return nil

}
