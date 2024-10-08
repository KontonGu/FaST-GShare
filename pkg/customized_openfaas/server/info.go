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

package server

import (
	"encoding/json"
	"net/http"

	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/version"
	"github.com/openfaas/faas-provider/types"
	klog "k8s.io/klog/v2"
)

func makeInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		sha, release := version.GetReleaseInfo()

		info := types.ProviderInfo{
			Orchestration: "kubernetes",
			Name:          "fastgshare-operator",
			Version: &types.VersionInfo{
				SHA:     sha,
				Release: release,
			},
		}

		infoBytes, err := json.Marshal(info)

		if err != nil {
			klog.Errorf("Failes to marshal info: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal info"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(infoBytes)
	}
}
