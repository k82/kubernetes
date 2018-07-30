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

package podschedulinggroup

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/apis/scheduling/validation"
)

// podSchedulingGroupStrategy implements verification logic for PodSchedulingGroup.
type podSchedulingGroupStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating PriorityClass objects.
var Strategy = podSchedulingGroupStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped returns true because all PriorityClasses are global.
func (podSchedulingGroupStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears the status of a PodSchedulingGroup before creation.
func (podSchedulingGroupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	pc := obj.(*scheduling.PodSchedulingGroup)
	pc.Generation = 1
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (podSchedulingGroupStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	_ = obj.(*scheduling.PodSchedulingGroup)
	_ = old.(*scheduling.PodSchedulingGroup)
}

// Validate validates a new PriorityClass.
func (podSchedulingGroupStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	psg := obj.(*scheduling.PodSchedulingGroup)
	return validation.ValidatePodSchedulingGroup(psg)
}

// Canonicalize normalizes the object after validation.
func (podSchedulingGroupStrategy) Canonicalize(obj runtime.Object) {}

// AllowCreateOnUpdate is false for PriorityClass; this means POST is needed to create one.
func (podSchedulingGroupStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (podSchedulingGroupStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidatePodSchedulingGroupUpdate(obj.(*scheduling.PodSchedulingGroup),
		old.(*scheduling.PodSchedulingGroup))
}

// AllowUnconditionalUpdate is the default update policy for PriorityClass objects.
func (podSchedulingGroupStrategy) AllowUnconditionalUpdate() bool {
	return true
}
