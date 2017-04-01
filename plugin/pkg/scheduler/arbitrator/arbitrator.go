package arbitrator

import (
	"k8s.io/kubernetes/pkg/api/v1"
	"fmt"
)

type Arbitrator struct {
	ResourceCache *ResourceCache
}

func NewArbitrator() *Arbitrator {
	rc := NewResourceCache()
	rm := NewResourceManager(rc)

	rm.Run()

	return &Arbitrator{
		ResourceCache: rc,
	}
}

func (a *Arbitrator) Admit(pod *v1.Pod) error {
	if alloc, found := a.ResourceCache.GetAllocation(pod.Namespace); found {
		req := GetPodRequest(pod)
		allocated := alloc.Allocated
		deserved := alloc.Deserved

		if allocated.Cpu().MilliValue() + req.Cpu().MilliValue() < deserved.Cpu().MilliValue() {
			return fmt.Errorf("cpu admit failed: request %d, allocated %d, deserved %d",
				allocated.Cpu().MilliValue(), req.Cpu().MilliValue(), deserved.Cpu().MilliValue())
		}

		if allocated.Memory().Value() + req.Memory().Value() < deserved.Memory().Value() {
			return fmt.Errorf("mem admit failed: request %d, allocated %d, deserved %d",
				allocated.Memory().Value(), req.Memory().Value(), deserved.Memory().Value())
		}
		return nil
	}

	return fmt.Errorf("failed to find allocation for %v/%v", pod.Namespace, pod.Name)
}
