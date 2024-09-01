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
	"io"
	"net/http"

	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func makeDeleteHandler(defaultNamespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace
		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		if namespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := io.ReadAll(r.Body)
		request := types.DeleteFunctionRequest{}

		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if len(request.FunctionName) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "No fastpod provided")
			return
		}

		// TODO: delete all FaSTPods related to the function name
		err = client.FastgshareV1().FaSTPods(lookupNamespace).
			Delete(r.Context(), request.FunctionName, metav1.DeleteOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			klog.Errorf("FaSTPod = %s delete error: %v", request.FunctionName, err)
		}
	}
}
