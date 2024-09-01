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
	"encoding/json"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	klog "k8s.io/klog/v2"
)

const (
	annotationFunctionSpec = "fast-gshare.fastpod.spec"
)

func makeAnnotations(fastpod *fastpodv1.FaSTPod) map[string]string {
	annotations := make(map[string]string)

	// disable scraping since the watchdog doesn't expose a metrics endpoint
	annotations["prometheus.io.scrape"] = "false"

	// copy function annotations
	if fastpod.Annotations != nil {
		for k, v := range fastpod.Annotations {
			annotations[k] = v
		}
	}

	// save function spec in deployment annotations
	// used to detect changes in function spec
	specJSON, err := json.Marshal(fastpod.Spec)
	if err != nil {
		klog.Errorf("Failed to marshal function spec: %s", err.Error())
		return annotations
	}

	annotations[annotationFunctionSpec] = string(specJSON)
	return annotations
}
