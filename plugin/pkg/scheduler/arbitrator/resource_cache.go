package arbitrator

import (
	"sync"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api/v1"
)

type ResourceCache struct {
	sync.Mutex

	Requests    map[string]*Consumer
	Allocations map[string]*Allocation
	Total       v1.ResourceList
}

func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		Requests: make(map[string]*Consumer),
		Allocations: make(map[string]*Allocation),
		Total: make(v1.ResourceList),
	}
}

func (rc *ResourceCache) addRequest(id string, reqs v1.ResourceList) {
	cons, found := rc.Requests[id]
	if !found {
		cons = NewConsumer(id)
		rc.Requests[id] = cons
	}

	for k, v := range reqs {
		cons.Request[k].Add(v)
	}
}

func (rc *ResourceCache) deleteRequest(id string, reqs v1.ResourceList) {
	cons, found := rc.Requests[id]
	if !found {
		cons = NewConsumer(id)
		rc.Requests[id] = cons
		glog.Warningf("Failed to delete request from a new consumer <%s>.", id)
		return
	}

	for k, v := range reqs {
		cons.Request[k].Sub(v)
	}
}

func (rc *ResourceCache) addAllocation(id string, reqs v1.ResourceList) {
	cons, found := rc.Requests[id]
	if !found {
		glog.Warningf("Failed to found Consumer <%s> when adding allocation.", id)
		return
	}

	alloc, found := rc.Allocations[id]
	if !found {
		alloc = NewAllocation(cons)
		rc.Requests[id] = alloc
	}

	for k, v := range reqs {
		alloc.Allocated[k].Add(v)
	}
}

func (rc *ResourceCache) deleteAllocation(id string, reqs v1.ResourceList) {
	cons, found := rc.Requests[id]
	if !found {
		glog.Warningf("Failed to found Consumer <%s> when deleting allocation.", id)
		return
	}

	alloc, found := rc.Allocations[id]
	if !found {
		alloc = NewAllocation(cons)
		rc.Requests[id] = alloc
		glog.Warningf("Failed to delete request from a new consumer <%s>.", id)
		return
	}

	for k, v := range reqs {
		alloc.Allocated[k].Sub(v)
	}
}

func (rc *ResourceCache) addTotalResource(reqs v1.ResourceList) {
	for k, v := range reqs {
		rc.Total[k].Add(v)
	}
}

func (rc *ResourceCache) deleteTotalResource(reqs v1.ResourceList) {
	for k, v := range reqs {
		rc.Total[k].Sub(v)
	}
}

func (rc *ResourceCache) AddPod(pod *v1.Pod) {
	rc.Lock()
	defer rc.Unlock()

	reqs := GetPodRequest(pod)
	rc.addRequest(pod.Namespace, reqs)
	if pod.Status.Phase == v1.PodRunning {
		rc.addAllocation(pod.Namespace, reqs)
	}

}

func (rc *ResourceCache) DeletePod(pod *v1.Pod) {
	rc.Lock()
	defer rc.Unlock()

	reqs := GetPodRequest(pod)
	rc.deleteRequest(pod.Namespace, reqs)
	if pod.Status.Phase == v1.PodRunning {
		rc.deleteAllocation(pod.Namespace, reqs)
	}
}

func (rc *ResourceCache) UpdatePod(old, pod *v1.Pod) {
	rc.Lock()
	defer rc.Unlock()

	oldReqs := GetPodRequest(old)
	rc.deleteRequest(pod.Namespace, oldReqs)
	if old.Status.Phase == v1.PodRunning {
		rc.deleteAllocation(pod.Namespace, oldReqs)
	}

	podReqs := GetPodRequest(pod)
	rc.addRequest(pod.Namespace, podReqs)
	if pod.Status.Phase == v1.PodRunning {
		rc.addAllocation(pod.Namespace, podReqs)
	}
}

func (rc *ResourceCache) AddNode(node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.addTotalResource(node.Status.Allocatable)
}

func (rc *ResourceCache) DeleteNode(node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.deleteTotalResource(node.Status.Allocatable)
}

func (rc *ResourceCache) Update(old, node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.deleteTotalResource(old.Status.Allocatable)
	rc.addTotalResource(node.Status.Allocatable)
}


func (rc *ResourceCache) GetAllocation(id string) (*Allocation, bool) {
	rc.Lock()
	rc.Unlock()

	// TODO: return a deepcopy of allocations
	return rc.Allocations[id]
}

func (rc *ResourceCache) GetSnapshot() (map[string]v1.ResourceList, map[string]v1.ResourceList, v1.ResourceList) {
	rc.Lock()
	rc.Unlock()

	
}