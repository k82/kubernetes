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

package internalversion

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling"
	scheme "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/scheme"
)

// PodSchedulingGroupsGetter has a method to return a PodSchedulingGroupInterface.
// A group's client should implement this interface.
type PodSchedulingGroupsGetter interface {
	PodSchedulingGroups(namespace string) PodSchedulingGroupInterface
}

// PodSchedulingGroupInterface has methods to work with PodSchedulingGroup resources.
type PodSchedulingGroupInterface interface {
	Create(*scheduling.PodSchedulingGroup) (*scheduling.PodSchedulingGroup, error)
	Update(*scheduling.PodSchedulingGroup) (*scheduling.PodSchedulingGroup, error)
	UpdateStatus(*scheduling.PodSchedulingGroup) (*scheduling.PodSchedulingGroup, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*scheduling.PodSchedulingGroup, error)
	List(opts v1.ListOptions) (*scheduling.PodSchedulingGroupList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *scheduling.PodSchedulingGroup, err error)
	PodSchedulingGroupExpansion
}

// podSchedulingGroups implements PodSchedulingGroupInterface
type podSchedulingGroups struct {
	client rest.Interface
	ns     string
}

// newPodSchedulingGroups returns a PodSchedulingGroups
func newPodSchedulingGroups(c *SchedulingClient, namespace string) *podSchedulingGroups {
	return &podSchedulingGroups{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the podSchedulingGroup, and returns the corresponding podSchedulingGroup object, and an error if there is any.
func (c *podSchedulingGroups) Get(name string, options v1.GetOptions) (result *scheduling.PodSchedulingGroup, err error) {
	result = &scheduling.PodSchedulingGroup{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PodSchedulingGroups that match those selectors.
func (c *podSchedulingGroups) List(opts v1.ListOptions) (result *scheduling.PodSchedulingGroupList, err error) {
	result = &scheduling.PodSchedulingGroupList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested podSchedulingGroups.
func (c *podSchedulingGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a podSchedulingGroup and creates it.  Returns the server's representation of the podSchedulingGroup, and an error, if there is any.
func (c *podSchedulingGroups) Create(podSchedulingGroup *scheduling.PodSchedulingGroup) (result *scheduling.PodSchedulingGroup, err error) {
	result = &scheduling.PodSchedulingGroup{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		Body(podSchedulingGroup).
		Do().
		Into(result)
	return
}

// Update takes the representation of a podSchedulingGroup and updates it. Returns the server's representation of the podSchedulingGroup, and an error, if there is any.
func (c *podSchedulingGroups) Update(podSchedulingGroup *scheduling.PodSchedulingGroup) (result *scheduling.PodSchedulingGroup, err error) {
	result = &scheduling.PodSchedulingGroup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		Name(podSchedulingGroup.Name).
		Body(podSchedulingGroup).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *podSchedulingGroups) UpdateStatus(podSchedulingGroup *scheduling.PodSchedulingGroup) (result *scheduling.PodSchedulingGroup, err error) {
	result = &scheduling.PodSchedulingGroup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		Name(podSchedulingGroup.Name).
		SubResource("status").
		Body(podSchedulingGroup).
		Do().
		Into(result)
	return
}

// Delete takes name of the podSchedulingGroup and deletes it. Returns an error if one occurs.
func (c *podSchedulingGroups) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *podSchedulingGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podschedulinggroups").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched podSchedulingGroup.
func (c *podSchedulingGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *scheduling.PodSchedulingGroup, err error) {
	result = &scheduling.PodSchedulingGroup{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("podschedulinggroups").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
