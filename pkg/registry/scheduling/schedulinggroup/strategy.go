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

package schedulinggroup

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/apis/scheduling/validation"
)

// schedulingGroupStrategy implements verification logic for PriorityClass.
type schedulingGroupStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating PriorityClass objects.
var Strategy = schedulingGroupStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped returns true because all PriorityClasses are global.
func (schedulingGroupStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears the status of a PriorityClass before creation.
func (schedulingGroupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	psg := obj.(*scheduling.PodSchedulingGroup)
	psg.Generation = 1
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (schedulingGroupStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	_ = obj.(*scheduling.PodSchedulingGroup)
	_ = old.(*scheduling.PodSchedulingGroup)
}

// Validate validates a new PriorityClass.
func (schedulingGroupStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	psg := obj.(*scheduling.PodSchedulingGroup)
	return validation.ValidatePodSchedulingGroup(psg)
}

// Canonicalize normalizes the object after validation.
func (schedulingGroupStrategy) Canonicalize(obj runtime.Object) {}

// AllowCreateOnUpdate is false for PodSchedulingGroup; this means POST is needed to create one.
func (schedulingGroupStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (schedulingGroupStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidatePodSchedulingGroupUpdate(
		obj.(*scheduling.PodSchedulingGroup), old.(*scheduling.PodSchedulingGroup))
}

// AllowUnconditionalUpdate is the default update policy for PodSchedulingGroup objects.
func (schedulingGroupStrategy) AllowUnconditionalUpdate() bool {
	return true
}

type schedulingGroupStatusStrategy struct {
	schedulingGroupStrategy
}

var StatusStrategy = schedulingGroupStatusStrategy{Strategy}

func (schedulingGroupStatusStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	_ = obj.(*scheduling.PodSchedulingGroup)
	// Nodes allow *all* fields, including status, to be set on create.
}

func (schedulingGroupStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newNode := obj.(*scheduling.PodSchedulingGroup)
	oldNode := old.(*scheduling.PodSchedulingGroup)
	newNode.Spec = oldNode.Spec
}

func (schedulingGroupStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidatePodSchedulingGroupUpdate(
		obj.(*scheduling.PodSchedulingGroup), old.(*scheduling.PodSchedulingGroup))
}

// Canonicalize normalizes the object after validation.
func (schedulingGroupStatusStrategy) Canonicalize(obj runtime.Object) {
}
