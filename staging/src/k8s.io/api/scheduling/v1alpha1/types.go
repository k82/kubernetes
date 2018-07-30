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

// Action represents the action that PodSchedulingGroup controller will take
// when event happen.
type Action string

// Event repreents the
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
	MinAvailable int `json:"minAvailable" protobuf:"bytes,1,opt,name=minAvailable"`
	// Policy defines the policy of PodSchedulingGroup lifecycle.
	// Default to 'Action: None, PodPhase: Failed'
	// +optional
	Policy []LifeCyclePolicy `json:"policy" protobuf:"bytes,2,opt,name=policy"`
}

// PodSchedulingGroupStatus represents the current state of a pod group.
type PodSchedulingGroupStatus struct {
	// The number of actively running pods.
	// +optional
	Running int32 `json:"running" protobuf:"bytes,1,opt,name=running"`
	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded" protobuf:"bytes,2,opt,name=succeeded"`
	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed" protobuf:"bytes,3,opt,name=failed"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSchedulingGroupList is a collection of pod group.
type PodSchedulingGroupList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of PodSchedulingGroup.
	Items []PodSchedulingGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}
