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
	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newService creates a new ClusterIP Service for a Function resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Function resource that 'owns' it.
func newService(fastpod *fastpodv1.FaSTPod) *corev1.Service {
	funcname, found := fastpod.Labels["fast_function"]
	if !found {
		funcname = fastpod.ObjectMeta.Name
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcname,
			Namespace: fastpod.Namespace,
			//todo here: do we need to modify annotaions also through fastpod
			Annotations: map[string]string{"prometheus.io.scrape": "false"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(fastpod, schema.GroupVersionKind{
					Group:   fastpodv1.SchemeGroupVersion.Group,
					Version: fastpodv1.SchemeGroupVersion.Version,
					Kind:    faasKind,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"fast_function": funcname},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     functionPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(functionPort),
					},
				},
			},
		},
	}
}
