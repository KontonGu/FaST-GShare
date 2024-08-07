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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	scheme "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// FaSTPodsGetter has a method to return a FaSTPodInterface.
// A group's client should implement this interface.
type FaSTPodsGetter interface {
	FaSTPods(namespace string) FaSTPodInterface
}

// FaSTPodInterface has methods to work with FaSTPod resources.
type FaSTPodInterface interface {
	Create(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.CreateOptions) (*v1.FaSTPod, error)
	Update(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.UpdateOptions) (*v1.FaSTPod, error)
	UpdateStatus(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.UpdateOptions) (*v1.FaSTPod, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.FaSTPod, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.FaSTPodList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.FaSTPod, err error)
	FaSTPodExpansion
}

// faSTPods implements FaSTPodInterface
type faSTPods struct {
	client rest.Interface
	ns     string
}

// newFaSTPods returns a FaSTPods
func newFaSTPods(c *FastgshareV1Client, namespace string) *faSTPods {
	return &faSTPods{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the faSTPod, and returns the corresponding faSTPod object, and an error if there is any.
func (c *faSTPods) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.FaSTPod, err error) {
	result = &v1.FaSTPod{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("fastpods").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of FaSTPods that match those selectors.
func (c *faSTPods) List(ctx context.Context, opts metav1.ListOptions) (result *v1.FaSTPodList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.FaSTPodList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("fastpods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested faSTPods.
func (c *faSTPods) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("fastpods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a faSTPod and creates it.  Returns the server's representation of the faSTPod, and an error, if there is any.
func (c *faSTPods) Create(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.CreateOptions) (result *v1.FaSTPod, err error) {
	result = &v1.FaSTPod{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("fastpods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(faSTPod).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a faSTPod and updates it. Returns the server's representation of the faSTPod, and an error, if there is any.
func (c *faSTPods) Update(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.UpdateOptions) (result *v1.FaSTPod, err error) {
	result = &v1.FaSTPod{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("fastpods").
		Name(faSTPod.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(faSTPod).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *faSTPods) UpdateStatus(ctx context.Context, faSTPod *v1.FaSTPod, opts metav1.UpdateOptions) (result *v1.FaSTPod, err error) {
	result = &v1.FaSTPod{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("fastpods").
		Name(faSTPod.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(faSTPod).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the faSTPod and deletes it. Returns an error if one occurs.
func (c *faSTPods) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("fastpods").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *faSTPods) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("fastpods").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched faSTPod.
func (c *faSTPods) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.FaSTPod, err error) {
	result = &v1.FaSTPod{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("fastpods").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
