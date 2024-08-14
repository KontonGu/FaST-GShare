/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.
Copyright (c) 2024 TUM - CAPS Cloud
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

package v1

import (
	"bytes"
	"math/rand"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// FaSTPod labels items' key name
const (
	FaSTGShareGPUQuotaRequest  = "fastgshare/gpu_quota_request"
	FaSTGShareGPUQuotaLimit    = "fastgshare/gpu_quota_limit"
	FaSTGShareGPUSMPartition   = "fastgshare/gpu_sm_partition"
	FaSTGShareGPUMemory        = "fastgshare/gpu_mem"
	FaSTGShareVGPUID           = "fastgshare/vgpu_id"
	FaSTGShareVGPUType         = "fastgshare/vgpu_type"
	FaSTGShareGPUsINfo         = "fastgshare/gpu_info"
	FaSTGShareNodeName         = "fastgshare/nodeName"
	FaSTGShareRole             = "fastgshare/role"
	FaSTGShareDummyPodName     = "fastgshare-vgpu"
	FaSTGShareDummyPodUUID     = "fastgshare/vgpu_uuid"
	OriginalNvidiaResourceName = "nvidia.com/gpu"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FaSTPod struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec FaSTPodSpec `json:"spec,omitempty"`
	// +optional
	Status FaSTPodStatus `json:"status,omitempty"`
}

type FaSTPodSpec struct {

	// +optional
	PodSpec corev1.PodSpec `json:"podSpec,omitempty"`

	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Selector is a label query over pods that should match the replica count.
	// Label keys and values that must match in order to be controlled by this replica set.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`
}

type FaSTPodStatus struct {
	// +optional
	PrewarmPool []*corev1.Pod `json:"prewarmPool,omitempty"` //list[*corev1.Pod]

	//PodObjectMeta *metav1.ObjectMeta

	// readyReplicas is the number of pods targeted by this ReplicaSet with a Ready Condition.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,4,opt,name=readyReplicas"`

	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty" protobuf:"varint,5,opt,name=availableReplicas"`

	// Replicas is the most recently oberved number of replicas.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	Replicas int32 `json:"replicas" protobuf:"varint,1,opt,name=replicas"`

	//maping from pod 2 boundDeviceID
	// +optional
	BoundDeviceIDs *map[string]string `json:"boundDeviceIDs,omitempty"`

	//+optional
	BoundDeviceType *map[string]string `json:"boundDeviceType,omitempty"`

	//pod to corresponding gpu client port
	GPUClientPort *map[string]int `json:"GPUClientPort,omitempty"`

	//TODOs: add replicas spec for faas
	// +optional
	Usage *map[string]FaSTPodUsage `json:"usage,omitempty"`

	// +optional
	Pod2Node *map[string]string `json:"pod2node,omitempty"`

	// +optional
	Node2Id []Scheded `json:"node2Id,omitempty"`

	//TODO: adding contitions?
}

type Scheded struct {
	Node string `json:"node,omitempty"`
	GPU  string `json:"gpu,omitempty"`
}

type FaSTPodUsage struct {
	GPU float64 `json:"gpu,omitempty"`

	TotalMemoryBytes float64 `json:"totalMemoryBytes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FaSTPod is a list of FaSTPod resources
type FaSTPodList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FaSTPod `json:"items"`
}

func (this FaSTPod) PrintInfo() {
	var buf bytes.Buffer
	buf.WriteString("\n================= FaSTPod ==================")
	buf.WriteString("\nname: ")
	buf.WriteString(this.ObjectMeta.Namespace)
	buf.WriteString("/")
	buf.WriteString(this.ObjectMeta.Name)
	buf.WriteString("\nannotation:\n\tfastgshare/gpu_request: ")
	buf.WriteString(this.ObjectMeta.Annotations["fastgshare/gpu_request"])
	buf.WriteString("\n\tGPUID: ")
	buf.WriteString(this.ObjectMeta.Annotations["fastgshare/GPUID"])
	buf.WriteString("\n\tBoundDeviceIs: ")
	//buf.WriteByte(this.Status.pod2BoundDeviceID)
	buf.WriteString("\n=============================================")
	klog.Info(buf.String())
}

func GenerateGPUID(n int) string {
	var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
