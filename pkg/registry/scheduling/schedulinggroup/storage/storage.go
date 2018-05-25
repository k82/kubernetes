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

package storage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/registry/scheduling/schedulinggroup"
)

// REST implements a RESTStorage for priority classes against etcd
type REST struct {
	*genericregistry.Store
}

type PodSchedulingGroupStorage struct {
	PodSchedulingGroup *REST
	Status             *StatusREST
}

// NewStorage returns a PodSchedulingGroupStorage object that will work against PodSchedulingGroupStorage.
func NewStorage(optsGetter generic.RESTOptionsGetter) PodSchedulingGroupStorage {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &scheduling.PodSchedulingGroup{} },
		NewListFunc:              func() runtime.Object { return &scheduling.PodSchedulingGroupList{} },
		DefaultQualifiedResource: scheduling.Resource("podschedulinggroup"),

		CreateStrategy: schedulinggroup.Strategy,
		UpdateStrategy: schedulinggroup.Strategy,
		DeleteStrategy: schedulinggroup.Strategy,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := *store
	statusStore.UpdateStrategy = schedulinggroup.StatusStrategy

	return PodSchedulingGroupStorage{
		PodSchedulingGroup: &REST{store},
		Status:             &StatusREST{store: &statusStore},
	}
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"psg"}
}

// StatusREST implements the REST endpoint for changing the status of a pod.
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &scheduling.PodSchedulingGroup{}
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation)
}
