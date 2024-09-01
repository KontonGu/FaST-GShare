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

package fastpod

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// FastpodResources is used to set GPU quota, partition, and memory limits and requests
type FastpodResources struct {
	GPUQuotaLimit   string `json:"gpuLimit"`
	GPUQuotaRequest string `json:"gpuRequest"`
	Memory          string `json:"memory"`
	SMPartition     string `json:"smParition"`
}

type FastpodStatus struct {
	// Name is the name of the function deployment
	Name string `json:"name"`
	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	Containers []corev1.Container `json:"containers,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FastpodResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FastpodResources `json:"requests,omitempty"`

	// InvocationCount count of invocations
	InvocationCount float64 `json:"invocationCount,omitempty"`

	// Replicas desired within the cluster
	Replicas uint64 `json:"replicas,omitempty"`

	// AvailableReplicas is the count of replicas ready to receive
	// invocations as reported by the faas-provider
	AvailableReplicas uint64 `json:"availableReplicas,omitempty"`

	// CreatedAt is the time read back from the faas backend's
	// data store for when the function or its container was created.
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// Usage represents CPU and RAM used by all of the
	// functions' replicas. Divide by AvailableReplicas for an
	// average value per replica.
	Usage *FastpodUsage `json:"usage,omitempty"`
}

type FastpodUsage struct {
	GPU              float64 `json:"gpu, omitempty"`
	TotalMemoryBytes float64 `json:"totalMemoryBytes, omitempty"`
}
