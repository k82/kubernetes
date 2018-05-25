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

package scheduling

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultPriorityWhenNoDefaultClassExists is used to set priority of pods
	// that do not specify any priority class and there is no priority class
	// marked as default.
	DefaultPriorityWhenNoDefaultClassExists = 0
	// HighestUserDefinablePriority is the highest priority for user defined priority classes. Priority values larger than 1 billion are reserved for Kubernetes system use.
	HighestUserDefinablePriority = int32(1000000000)
	// SystemCriticalPriority is the beginning of the range of priority values for critical system components.
	SystemCriticalPriority = 2 * HighestUserDefinablePriority
	// SystemPriorityClassPrefix is the prefix reserved for system priority class names. Other priority
	// classes are not allowed to start with this prefix.
	SystemPriorityClassPrefix = "system-"
	// NOTE: In order to avoid conflict of names with user-defined priority classes, all the names must
	// start with SystemPriorityClassPrefix.
	SystemClusterCritical = SystemPriorityClassPrefix + "cluster-critical"
	SystemNodeCritical    = SystemPriorityClassPrefix + "node-critical"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityClass defines the mapping from a priority class name to the priority
// integer value. The value can be any valid integer.
type PriorityClass struct {
	metav1.TypeMeta
	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta

	// The value of this priority class. This is the actual priority that pods
	// receive when they have the name of this class in their pod spec.
	Value int32

	// globalDefault specifies whether this PriorityClass should be considered as
	// the default priority for pods that do not have any priority class.
	// Only one PriorityClass can be marked as `globalDefault`. However, if more than
	// one PriorityClasses exists with their `globalDefault` field set to true,
	// the smallest value of such global default PriorityClasses will be used as the default priority.
	// +optional
	GlobalDefault bool

	// Description is an arbitrary string that usually provides guidelines on
	// when this priority class should be used.
	// +optional
	Description string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityClassList is a collection of priority classes.
type PriorityClassList struct {
	metav1.TypeMeta
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta

	// Items is the list of PriorityClasses.
	Items []PriorityClass
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroup is a collection of Pods.
type PodSchedulingGroup struct {
	metav1.TypeMeta
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PodSchedulingGroupSpec

	// Most recently observed status of the pod.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status PodSchedulingGroupStatus
}

type PodSchedulingGroupSpec struct {
	// Selector is a label query over pods that should be considered together by scheduler.
	// +optional
	Selector *metav1.LabelSelector

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string

	// The priority value. Various system components use this field to find the
	// priority of the PodSet. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32

	// The minimal available pods to run; the default value is nil.
	// +optional
	MinAvailable *int32
}

type PodSchedulingGroupStatus struct {
	// The number of pending pods.
	// +optional
	Pending int32

	// The number of actively running pods.
	// +optional
	Running int32

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32

	// Replicas is the number of selected pods.
	// +optional
	Replicas int32

	// The minimal available pods to run for this QueueJob
	// +optional
	MinAvailable int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSetList is a collection of PodSet.
type PodSchedulingGroupList struct {
	metav1.TypeMeta
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta

	// Items is the list of PodSchedulingGroup.
	Items []PodSchedulingGroup
}
