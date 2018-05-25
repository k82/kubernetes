/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityClass defines mapping from a priority class name to the priority
// integer value. The value can be any valid integer.
type PriorityClass struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// The value of this priority class. This is the actual priority that pods
	// receive when they have the name of this class in their pod spec.
	Value int32 `json:"value" protobuf:"bytes,2,opt,name=value"`

	// globalDefault specifies whether this PriorityClass should be considered as
	// the default priority for pods that do not have any priority class.
	// Only one PriorityClass can be marked as `globalDefault`. However, if more than
	// one PriorityClasses exists with their `globalDefault` field set to true,
	// the smallest value of such global default PriorityClasses will be used as the default priority.
	// +optional
	GlobalDefault bool `json:"globalDefault,omitempty" protobuf:"bytes,3,opt,name=globalDefault"`

	// description is an arbitrary string that usually provides guidelines on
	// when this priority class should be used.
	// +optional
	Description string `json:"description,omitempty" protobuf:"bytes,4,opt,name=description"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityClassList is a collection of priority classes.
type PriorityClassList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of PriorityClasses
	Items []PriorityClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroup is the configuration
type PodSchedulingGroup struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PodSchedulingGroupSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the pod.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status PodSchedulingGroupStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type PodSchedulingGroupSpec struct {
	// Selector is a label query over pods that should be considered together by scheduler.
	// +optional
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,2,opt,name=priorityClassName"`

	// The priority value. Various system components use this field to find the
	// priority of the QueueJob. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority" protobuf:"bytes,3,opt,name=priority"`

	// The minimal available pods to run for this QueueJob; the default value is nil.
	// +optional
	MinAvailable *int32 `json:"minAvailable" protobuf:"bytes,4,opt,name=minAvailable"`
}

type PodSchedulingGroupStatus struct {
	// The number of pending pods.
	// +optional
	Pending int32 `json:"pending" protobuf:"bytes,1,opt,name=pending"`

	// The number of actively running pods.
	// +optional
	Running int32 `json:"running" protobuf:"bytes,2,opt,name=running"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded" protobuf:"bytes,3,opt,name=succeeded"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed" protobuf:"bytes,4,opt,name=failed"`

	// Replicas is the number of desired replicas.
	// +optional
	Replicas int32 `json:"replicas" protobuf:"bytes,5,opt,name=replicas"`

	// The minimal available pods to run for this QueueJob
	// +optional
	MinAvailable int32 `json:"minAvailable" protobuf:"bytes,6,opt,name=minAvailable"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroup is a collection of priority classes.
type PodSchedulingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of PodSchedulingGroup
	Items []PodSchedulingGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}
