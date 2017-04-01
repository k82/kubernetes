package arbitrator

import "k8s.io/kubernetes/plugin/pkg/scheduler/arbitrator/allocator"

type ResourceManager struct {
	ResourceCache *ResourceCache
	Allocate allocator.Allocate
}

func NewResourceManager(rc *ResourceCache) *ResourceManager {
	return &ResourceManager{
		ResourceCache: rc,
		Allocate: allocator.DRFAllocate,
	}
}

func (rm *ResourceManager) Run() {
	cons, alloc, total := rm.ResourceCache.GetSnapshot()

	newAlloc := rm.Allocate(cons, alloc, total)

	rm.ResourceCache.SetAllocations(newAlloc)
}