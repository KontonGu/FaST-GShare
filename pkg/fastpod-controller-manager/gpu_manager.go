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

import "container/list"

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
	VGPUID2GPU map[string]*GPUDevInfo
	// The port allocator for gpu clients, fast schedulers, and the configurator
	DaemonPortAlloc *bitmap.RRBitmap
}
