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

package k8s

import (
	"sync"
	"time"
)

type PodInfo struct {
	PodName          string
	PodIp            string
	ServiceName      string
	AvgResponseTime  time.Duration
	LastResponseTime time.Duration
	TotalInvoke      int32
	LastInvoke       time.Time
	Rate             float32
	//
	RateChange   ChangeType
	PossiTimeout bool
	Timeout      bool
}

type FaSTPodInfo struct {
	//PodInfos map[string]PodInfo
	ScaleDown   bool
	TotalInvoke int32
	//todo make thread safe
	Lock sync.RWMutex
	//Lock sync.Mutex
}

// may not defined here
type ContainerPool struct {
	shrName     string
	containerId string
	podName     string
}

type ChangeType int

const (
	Inc ChangeType = 2
	Dec ChangeType = 0
	Sta ChangeType = 1
)
