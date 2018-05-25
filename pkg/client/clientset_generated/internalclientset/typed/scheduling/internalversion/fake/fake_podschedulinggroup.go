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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling"
)

// FakePodSchedulingGroups implements PodSchedulingGroupInterface
type FakePodSchedulingGroups struct {
	Fake *FakeScheduling
	ns   string
}

var podschedulinggroupsResource = schema.GroupVersionResource{Group: "scheduling.k8s.io", Version: "", Resource: "podschedulinggroups"}

var podschedulinggroupsKind = schema.GroupVersionKind{Group: "scheduling.k8s.io", Version: "", Kind: "PodSchedulingGroup"}

// Get takes name of the podSchedulingGroup, and returns the corresponding podSchedulingGroup object, and an error if there is any.
func (c *FakePodSchedulingGroups) Get(name string, options v1.GetOptions) (result *scheduling.PodSchedulingGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podschedulinggroupsResource, c.ns, name), &scheduling.PodSchedulingGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*scheduling.PodSchedulingGroup), err
}

// List takes label and field selectors, and returns the list of PodSchedulingGroups that match those selectors.
func (c *FakePodSchedulingGroups) List(opts v1.ListOptions) (result *scheduling.PodSchedulingGroupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podschedulinggroupsResource, podschedulinggroupsKind, c.ns, opts), &scheduling.PodSchedulingGroupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &scheduling.PodSchedulingGroupList{}
	for _, item := range obj.(*scheduling.PodSchedulingGroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podSchedulingGroups.
func (c *FakePodSchedulingGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podschedulinggroupsResource, c.ns, opts))

}

// Create takes the representation of a podSchedulingGroup and creates it.  Returns the server's representation of the podSchedulingGroup, and an error, if there is any.
func (c *FakePodSchedulingGroups) Create(podSchedulingGroup *scheduling.PodSchedulingGroup) (result *scheduling.PodSchedulingGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podschedulinggroupsResource, c.ns, podSchedulingGroup), &scheduling.PodSchedulingGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*scheduling.PodSchedulingGroup), err
}

// Update takes the representation of a podSchedulingGroup and updates it. Returns the server's representation of the podSchedulingGroup, and an error, if there is any.
func (c *FakePodSchedulingGroups) Update(podSchedulingGroup *scheduling.PodSchedulingGroup) (result *scheduling.PodSchedulingGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podschedulinggroupsResource, c.ns, podSchedulingGroup), &scheduling.PodSchedulingGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*scheduling.PodSchedulingGroup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePodSchedulingGroups) UpdateStatus(podSchedulingGroup *scheduling.PodSchedulingGroup) (*scheduling.PodSchedulingGroup, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podschedulinggroupsResource, "status", c.ns, podSchedulingGroup), &scheduling.PodSchedulingGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*scheduling.PodSchedulingGroup), err
}

// Delete takes name of the podSchedulingGroup and deletes it. Returns an error if one occurs.
func (c *FakePodSchedulingGroups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(podschedulinggroupsResource, c.ns, name), &scheduling.PodSchedulingGroup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodSchedulingGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podschedulinggroupsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &scheduling.PodSchedulingGroupList{})
	return err
}

// Patch applies the patch and returns the patched podSchedulingGroup.
func (c *FakePodSchedulingGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *scheduling.PodSchedulingGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podschedulinggroupsResource, c.ns, name, data, subresources...), &scheduling.PodSchedulingGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*scheduling.PodSchedulingGroup), err
}
