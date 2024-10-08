//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaSTPod) DeepCopyInto(out *FaSTPod) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaSTPod.
func (in *FaSTPod) DeepCopy() *FaSTPod {
	if in == nil {
		return nil
	}
	out := new(FaSTPod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FaSTPod) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaSTPodList) DeepCopyInto(out *FaSTPodList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FaSTPod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaSTPodList.
func (in *FaSTPodList) DeepCopy() *FaSTPodList {
	if in == nil {
		return nil
	}
	out := new(FaSTPodList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FaSTPodList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaSTPodSpec) DeepCopyInto(out *FaSTPodSpec) {
	*out = *in
	in.PodSpec.DeepCopyInto(&out.PodSpec)
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaSTPodSpec.
func (in *FaSTPodSpec) DeepCopy() *FaSTPodSpec {
	if in == nil {
		return nil
	}
	out := new(FaSTPodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaSTPodStatus) DeepCopyInto(out *FaSTPodStatus) {
	*out = *in
	if in.PrewarmPool != nil {
		in, out := &in.PrewarmPool, &out.PrewarmPool
		*out = make([]*corev1.Pod, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.Pod)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.BoundDeviceIDs != nil {
		in, out := &in.BoundDeviceIDs, &out.BoundDeviceIDs
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.BoundDeviceType != nil {
		in, out := &in.BoundDeviceType, &out.BoundDeviceType
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.GPUClientPort != nil {
		in, out := &in.GPUClientPort, &out.GPUClientPort
		*out = new(map[string]int)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]int, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.Usage != nil {
		in, out := &in.Usage, &out.Usage
		*out = new(map[string]FaSTPodUsage)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]FaSTPodUsage, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.Pod2Node != nil {
		in, out := &in.Pod2Node, &out.Pod2Node
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.Node2Id != nil {
		in, out := &in.Node2Id, &out.Node2Id
		*out = make([]Scheded, len(*in))
		copy(*out, *in)
	}
	if in.ResourceConfig != nil {
		in, out := &in.ResourceConfig, &out.ResourceConfig
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaSTPodStatus.
func (in *FaSTPodStatus) DeepCopy() *FaSTPodStatus {
	if in == nil {
		return nil
	}
	out := new(FaSTPodStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaSTPodUsage) DeepCopyInto(out *FaSTPodUsage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaSTPodUsage.
func (in *FaSTPodUsage) DeepCopy() *FaSTPodUsage {
	if in == nil {
		return nil
	}
	out := new(FaSTPodUsage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scheded) DeepCopyInto(out *Scheded) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scheded.
func (in *Scheded) DeepCopy() *Scheded {
	if in == nil {
		return nil
	}
	out := new(Scheded)
	in.DeepCopyInto(out)
	return out
}
