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
	"fmt"
	"io"
	"net/http"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	faasfastpod "github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/fastpod"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klog "k8s.io/klog/v2"
)

func makeApplyHandler(defaultNamespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := io.ReadAll(r.Body)

		req := faasfastpod.FastpodDeployment{}

		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		klog.Infof("Deployment request for : %s\n", req.Service)

		namespace := defaultNamespace
		if len(req.Namespace) > 0 {
			namespace = req.Namespace
		}

		opts := metav1.GetOptions{}
		got, err := client.FastgshareV1().FaSTPods(namespace).Get(r.Context(), req.Service, opts)
		miss := false

		if err != nil {
			if errors.IsNotFound(err) {
				miss = true
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		// Sometimes there was a non-nil "got" value when the miss was
		// true.
		if miss == false && got != nil {
			updated := got.DeepCopy()

			klog.Infof("Updating %s/n", updated.ObjectMeta.Name)

			updated.Spec = toFastpod(req, namespace).Spec
			updated.Labels = toFastpod(req, namespace).Labels

			if _, err = client.FastgshareV1().FaSTPods(namespace).Update(r.Context(), updated, metav1.UpdateOptions{}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error updating fastpod : %s", err.Error())))
				return
			}
		} else {
			//TODO: finishing the fastpod deployment
			newFastPod := toFastpod(req, namespace)

			if _, err = client.FastgshareV1().FaSTPods(namespace).
				Create(r.Context(), &newFastPod, metav1.CreateOptions{}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error creating fastpod: %s", err.Error())))
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func toFastpod(req faasfastpod.FastpodDeployment, namespace string) fastpodv1.FaSTPod {
	fastpod := fastpodv1.FaSTPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Service,
			Namespace:   namespace,
			Annotations: req.Annotations,
			Labels:      *req.Labels,
		},
		Spec: fastpodv1.FaSTPodSpec{
			PodSpec: corev1.PodSpec{
				Containers: req.Containers,
			},
		},
	}

	return fastpod
}
