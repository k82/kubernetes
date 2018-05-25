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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

func TestPodSchedulingGroupStrategy(t *testing.T) {
	p1 := int32(1)
	p2 := int32(2)

	ctx := genericapirequest.NewDefaultContext()
	if !Strategy.NamespaceScoped() {
		t.Errorf("PodSchedulingGroup must be namespace scoped")
	}
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("PodSchedulingGroup should not allow create on update")
	}

	psg := &scheduling.PodSchedulingGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "valid-class",
		},
		Spec: scheduling.PodSchedulingGroupSpec{
			Priority: &p1,
		},
	}

	Strategy.PrepareForCreate(ctx, psg)

	errs := Strategy.Validate(ctx, psg)
	if len(errs) != 0 {
		t.Errorf("unexpected error validating %v", errs)
	}

	newPSG := &scheduling.PodSchedulingGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "valid-class-2",
			ResourceVersion: "4",
		},
		Spec: scheduling.PodSchedulingGroupSpec{
			Priority: &p2,
		},
	}

	Strategy.PrepareForUpdate(ctx, newPSG, psg)

	errs = Strategy.ValidateUpdate(ctx, newPSG, psg)
	if len(errs) == 0 {
		t.Errorf("Expected a validation error")
	}
}
