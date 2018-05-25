/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetDefaults_PodSchedulingGroupSpec(spec *scheduling.PodSchedulingGroupSpec) {
	if spec.MinAvailable == nil {
		one := int32(1)
		spec.MinAvailable = &one
	}

	if spec.Selector == nil {
		nothing := v1.LabelSelector{}
		spec.Selector = &nothing
	}
}
