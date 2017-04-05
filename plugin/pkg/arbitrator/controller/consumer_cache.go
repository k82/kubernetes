package controller

import (
	"sync"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/arbitrator/api"
)

type ConsumerCache struct {
	sync.Mutex

	Requests    map[string]*api.Consumer

	Nodes       map[string]*api.NodeInfo
}

func NewResourceCache() *ConsumerCache {
	return &ConsumerCache{
		Requests: make(map[string]*api.Consumer),
		Allocations: make(map[string]*api.Allocation),
		Total: make(v1.ResourceList),
	}
}

func (rc *ConsumerCache) addRequest(id string, reqs v1.ResourceList) {
	cons, found := rc.Requests[id]
	if !found {
		cons = NewConsumer(id)
		rc.Requests[id] = cons
	}

	for k, v := range reqs {
		cons.Request[k].Add(v)
	}
}

func (rc *ConsumerCache) deleteRequest(id string, reqs v1.ResourceList) {
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

func (rc *ConsumerCache) addAllocation(id string, reqs v1.ResourceList) {
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

func (rc *ConsumerCache) deleteAllocation(id string, reqs v1.ResourceList) {
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

func (rc *ConsumerCache) addTotalResource(reqs v1.ResourceList) {
	for k, v := range reqs {
		rc.Total[k].Add(v)
	}
}

func (rc *ConsumerCache) deleteTotalResource(reqs v1.ResourceList) {
	for k, v := range reqs {
		rc.Total[k].Sub(v)
	}
}

func (rc *ConsumerCache) AddPod(pod *v1.Pod) {
	rc.Lock()
	defer rc.Unlock()

	reqs := GetPodRequest(pod)
	rc.addRequest(pod.Namespace, reqs)
	if pod.Status.Phase == v1.PodRunning {
		rc.addAllocation(pod.Namespace, reqs)
	}

}

func (rc *ConsumerCache) DeletePod(pod *v1.Pod) {
	rc.Lock()
	defer rc.Unlock()

	reqs := GetPodRequest(pod)
	rc.deleteRequest(pod.Namespace, reqs)
	if pod.Status.Phase == v1.PodRunning {
		rc.deleteAllocation(pod.Namespace, reqs)
	}
}

func (rc *ConsumerCache) UpdatePod(old, pod *v1.Pod) {
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

func (rc *ConsumerCache) AddNode(node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.addTotalResource(node.Status.Allocatable)
}

func (rc *ConsumerCache) DeleteNode(node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.deleteTotalResource(node.Status.Allocatable)
}

func (rc *ConsumerCache) Update(old, node *v1.Node) {
	rc.Lock()
	rc.Unlock()

	rc.deleteTotalResource(old.Status.Allocatable)
	rc.addTotalResource(node.Status.Allocatable)
}