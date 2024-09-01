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

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klog "k8s.io/klog/v2"
)

func makeListHandler(defaultNamespace string,
	client clientset.Interface,
	fastpodLister listers.FaSTPodLister) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		q := r.URL.Query()

		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 && namespace != "fast-gshare" {
			lookupNamespace = namespace
		}

		if lookupNamespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system", http.StatusUnauthorized)
			return
		}

		opts := metav1.ListOptions{}
		res, err := client.FastgshareV1().FaSTPods(lookupNamespace).List(r.Context(), opts)
		//todo here
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			klog.Errorf("Fastpod listing error: %v", err)
		}

		fastpodsStatus := []fastpodv1.FaSTPodStatus{}

		for _, item := range res.Items {
			//desiredReplicas, availableReplicas, err := getReplicas(item.Name, lookupNamespace, deploymentLister)
			if err != nil {
				klog.Warningf("Function listing getReplicas error: %v", err)
			}

			faststatus := item.Status
			fastpodsStatus = append(fastpodsStatus, faststatus)
		}

		fastpodBytes, err := json.Marshal(fastpodsStatus)
		if err != nil {
			klog.Errorf("Failed to marshal fastpods: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal fastpods"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fastpodBytes)
	}
}
