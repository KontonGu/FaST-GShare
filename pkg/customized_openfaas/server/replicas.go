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

	"io"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	faasfastpod "github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/fastpod"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/k8s"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klog "k8s.io/klog/v2"
)

func makeReplicaReader(defaultNamespace string, client clientset.Interface, lister listers.FaSTPodLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fastpodName := vars["name"]

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		opts := metav1.GetOptions{}

		k8sfstp, err := client.FastgshareV1().FaSTPods(lookupNamespace).Get(r.Context(), fastpodName, opts)

		if err != nil || k8sfstp == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		desiredReplicas, availableReplicas, err := getReplicas(fastpodName, lookupNamespace, lister)

		if err != nil {
			klog.Warningf("Fastpod replica reader error: %v", err)
		}

		result := toFastpodStatus(*k8sfstp)

		result.AvailableReplicas = availableReplicas
		result.Replicas = desiredReplicas

		res, err := json.Marshal(result)
		if err != nil {
			klog.Errorf("Failed to marshal fastpod status: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal fastpod status"))
			return
		}

		w.Header().Set("Content-Type", "applications/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func getReplicas(fastpodName string, namespace string, lister listers.FaSTPodLister) (uint64, uint64, error) {
	fstp, err := lister.FaSTPods(namespace).Get(fastpodName)
	if err != nil {
		return 0, 0, err
	}

	desiredReplicas := uint64(fstp.Status.Replicas)
	availableReplicas := uint64(fstp.Status.AvailableReplicas)

	return desiredReplicas, availableReplicas, nil
}

func toFastpodStatus(item fastpodv1.FaSTPod) faasfastpod.FastpodStatus {
	status := faasfastpod.FastpodStatus{
		Labels:            &item.Labels,
		Annotations:       &item.Annotations,
		Name:              item.Name,
		Containers:        item.Spec.PodSpec.Containers,
		CreatedAt:         item.CreationTimestamp.Time,
		AvailableReplicas: uint64(item.Status.AvailableReplicas),
		Replicas:          uint64(item.Status.Replicas),
	}
	//TODO, we should specify limit & request here.
	return status
}

func makeReplicaHandler(defaultNamespace string, client clientset.Interface, functionlookup *k8s.FunctionLookup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fstpDepName := vars["name"]

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		if lookupNamespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		req := types.ScaleServiceRequest{}

		if r.Body != nil {
			defer r.Body.Close()
			bytesIn, _ := io.ReadAll(r.Body)

			if err := json.Unmarshal(bytesIn, &req); err != nil {
				klog.Errorf("Function %s replica invalid JSON: %v", fstpDepName, err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		}

		opts := metav1.GetOptions{}
		//need to update here!
		fstp, err := client.FastgshareV1().FaSTPods(lookupNamespace).Get(r.Context(), fstpDepName, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			klog.Errorf("Fastpod %s get error: %v", fstpDepName, err)
			return
		}
		fstpCopy := fstp.DeepCopy()

		//TODO: what could done better here?
		// Currently the server interface is not responsible for auto-scaling
		// The FaST-GShare-Autoscaler is for this function.
		replica := fstp.Spec.Replicas
		if *replica >= int32(req.Replicas) {
			//go functionlookup.ScaleDown(fstpDepName)
		} else {
			//go functionlookup.ScaleUp(fstpDepName)
		}

		klog.Infof("Current replica is %d", *replica)

		fstpCopy.Spec.Replicas = int32p(int32(req.Replicas))
		klog.Infof("Replicas from request = %d.", req.Replicas)
		updatedFstp, err := client.FastgshareV1().FaSTPods(lookupNamespace).Update(r.Context(), fstpCopy, metav1.UpdateOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			klog.Errorf("Fastpod %s update error %v", fstpDepName, err)
			return
		}

		if *updatedFstp.Spec.Replicas == int32(req.Replicas) {
			klog.Infof("Fastpod %s replica updated to %v", fstpDepName, req.Replicas)
			w.WriteHeader(http.StatusAccepted)
		} else {
			klog.Infof("Fastpod %s with replica %i failed updated to %v replicas", updatedFstp.Name, *updatedFstp.Spec.Replicas, req.Replicas)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func int32p(i int32) *int32 {
	return &i
}
