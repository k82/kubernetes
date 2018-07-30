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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

type Action string
type Event string

const (
	RestartAction Action = "restart"
	NoneAction    Action = "none"

	PodFailedEvent     Event = "PodFailed"
	UnschedulableEvent Event = "Unschedulable"
)

// LifeCyclePolicy represents the lifecycle policy of PodSchedulingGroup
// according to Pod's phase.
type LifeCyclePolicy struct {
	// The action that will be taken to the PodSchedulingGroup according to
	// Pod's phase. One of "Restart", "None".
	// Default to None.
	Action Action
	// The phase of pod; the controller takes actions according to this
	// pod's phase. One of "PodFailed", "Unschedulable".
	Event Event
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroup defines the scheduling requirement of a pod group
type PodSchedulingGroup struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   PodSchedulingGroupTemplate
	Status PodSchedulingGroupStatus
}

// PodSchedulingGroupTemplate represents the template of a pod group.
type PodSchedulingGroupTemplate struct {
	// MinAvailable defines the minimal available tasks to run the Job;
	// if there's not enough resources to start all tasks, the scheduler
	// will not start anyone.
	MinAvailable int
	// Policy defines the policy of PodSchedulingGroup lifecycle.
	// Default to 'Action: None, PodPhase: Failed'
	// +optional
	Policy []LifeCyclePolicy
}

// PodSchedulingGroupStatus represents the current state of a pod group.
type PodSchedulingGroupStatus struct {
	// The number of actively running pods.
	// +optional
	Running int32
	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32
	// The number of pods which reached phase Failed.
	// +optional
	Failed int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroupList is a collection of pod group.
type PodSchedulingGroupList struct {
	metav1.TypeMeta
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta

	// Items is the list of PodSchedulingGroup.
	Items []PodSchedulingGroup
}
