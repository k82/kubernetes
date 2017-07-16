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

package node

import (
	"sync"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
	nodeutil "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	corelisters "k8s.io/kubernetes/pkg/client/listers/core/v1"
	utilnode "k8s.io/kubernetes/pkg/util/node"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/controller"
	v1helper "k8s.io/kubernetes/pkg/api/v1/helper"

	"github.com/golang/glog"
	"fmt"
)

const (
	// controls how often NodeController will try to evict Pods from non-responsive Nodes.
	nodeEvictionPeriod = 100 * time.Millisecond
	// Burst value for all eviction rate limiters
	evictionRateLimiterBurst = 1
)

type podEvictor interface {
	// Managing eviction of nodes:
	// When we delete pods off a node, if the node was not empty at the time we then
	// queue an eviction watcher. If we hit an error, retry deletion.
	Run()

	CancelPodEviction(node *v1.Node) error

	EvictPods(node *v1.Node, observed *v1.NodeCondition, timestamp metav1.Time) error

	AddPodEvictor(zone string) error

	// SwapLimiter safely swaps current limiter for this queue with the passed one if capacities or qps's differ.
	SwapEvictorLimiter(zone string, newQPS float32)
}

type taintBasedEvictor struct {
	// Lock to access evictor workers
	evictorLock        sync.Mutex
	zoneTainer         map[string]*RateLimitedTimedQueue
	nodeLister         corelisters.NodeLister
	kubeClient         clientset.Interface
	evictionLimiterQPS float32
}

func (t *taintBasedEvictor) Run() {
	go wait.Until(t.doTaintingPass, nodeEvictionPeriod, wait.NeverStop)
}

func (t *taintBasedEvictor) CancelPodEviction(node *v1.Node) error {
	t.evictorLock.Lock()
	defer t.evictorLock.Unlock()

	err := controller.RemoveTaintOffNode(t.kubeClient, node.Name, UnreachableTaintTemplate, node)
	if err != nil {
		glog.Errorf("Failed to remove taint from node %v: %v", node.Name, err)
		return err
	}
	err = controller.RemoveTaintOffNode(t.kubeClient, node.Name, NotReadyTaintTemplate, node)
	if err != nil {
		glog.Errorf("Failed to remove taint from node %v: %v", node.Name, err)
		return  err
	}

	t.zoneTainer[utilnode.GetZoneKey(node)].Remove(node.Name)

	return nil
}

func (t *taintBasedEvictor) EvictPods(node *v1.Node, observed *v1.NodeCondition, timestamp metav1.Time) error {
	t.evictorLock.Lock()
	defer t.evictorLock.Unlock()

	if observed.Status == v1.ConditionFalse {
		// We want to update the taint straight away if Node is already tainted with the UnreachableTaint
		if v1helper.TaintExists(node.Spec.Taints, NotReadyTaintTemplate) {
			taintToAdd := *UnreachableTaintTemplate
			if !swapNodeControllerTaint(t.kubeClient, &taintToAdd, NotReadyTaintTemplate, node) {
				glog.Errorf("Failed to instantly swap UnreachableTaint to NotReadyTaint. Will try again in the next cycle.")
			}
		} else if t.zoneTainer[utilnode.GetZoneKey(node)].Add(node.Name, string(node.UID)) {
			glog.V(2).Infof("Node %v is unresponsive as of %v. Adding it to the Taint queue.",
				node.Name,
				metav1.Now(),
			)
		}
	} else if observed.Status == v1.ConditionUnknown {
		// We want to update the taint straight away if Node is already tainted with the UnreachableTaint
		if v1helper.TaintExists(node.Spec.Taints, NotReadyTaintTemplate) {
			taintToAdd := *UnreachableTaintTemplate
			if !swapNodeControllerTaint(t.kubeClient, &taintToAdd, NotReadyTaintTemplate, node) {
				glog.Errorf("Failed to instantly swap UnreachableTaint to NotReadyTaint. Will try again in the next cycle.")
			}
		} else if t.zoneTainer[utilnode.GetZoneKey(node)].Add(node.Name, string(node.UID)) {
			glog.V(2).Infof("Node %v is unresponsive as of %v. Adding it to the Taint queue.",
				node.Name,
				metav1.Now(),
			)
		}
	}

	return nil
}


func (t *taintBasedEvictor) SwapEvictorLimiter(zone string, newQPS float32) {
	t.zoneTainer[zone].SwapLimiter(newQPS)
}

func (t *taintBasedEvictor) AddPodEvictor(zone string) error {
	t.evictorLock.Lock()
	defer t.evictorLock.Unlock()

	t.zoneTainer[zone] =
		NewRateLimitedTimedQueue(
			flowcontrol.NewTokenBucketRateLimiter(t.evictionLimiterQPS, evictionRateLimiterBurst))
	return nil
}

func (t *taintBasedEvictor) doTaintingPass() {
	t.evictorLock.Lock()
	defer t.evictorLock.Unlock()

	for k := range t.zoneTainer {
		// Function should return 'false' and a time after which it should be retried, or 'true' if it shouldn't (it succeeded).
		t.zoneTainer[k].Try(func(value TimedValue) (bool, time.Duration) {
			node, err := t.nodeLister.Get(value.Value)
			if apierrors.IsNotFound(err) {
				glog.Warningf("Node %v no longer present in nodeLister!", value.Value)
				return true, 0
			} else if err != nil {
				glog.Warningf("Failed to get Node %v from the nodeLister: %v", value.Value, err)
				// retry in 50 millisecond
				return false, 50 * time.Millisecond
			} else {
				zone := utilnode.GetZoneKey(node)
				EvictionsNumber.WithLabelValues(zone).Inc()
			}
			_, condition := nodeutil.GetNodeCondition(&node.Status, v1.NodeReady)
			// Because we want to mimic NodeStatus.Condition["Ready"] we make "unreachable" and "not ready" taints mutually exclusive.
			taintToAdd := v1.Taint{}
			oppositeTaint := v1.Taint{}
			if condition.Status == v1.ConditionFalse {
				taintToAdd = *NotReadyTaintTemplate
				oppositeTaint = *UnreachableTaintTemplate
			} else if condition.Status == v1.ConditionUnknown {
				taintToAdd = *UnreachableTaintTemplate
				oppositeTaint = *NotReadyTaintTemplate
			} else {
				// It seems that the Node is ready again, so there's no need to taint it.
				glog.V(4).Infof("Node %v was in a taint queue, but it's ready now. Ignoring taint request.", value.Value)
				return true, 0
			}

			return swapNodeControllerTaint(t.kubeClient, &taintToAdd, &oppositeTaint, node), 0
		})
	}
}

type defaultPodEvictor struct {
	evictorLock        sync.Mutex
	zonePodEvictor         map[string]*RateLimitedTimedQueue
	nodeLister         corelisters.NodeLister
	kubeClient         clientset.Interface
	evictionLimiterQPS float32
}

func (d *defaultPodEvictor) Run() {

}
func (d *defaultPodEvictor) doEvictionPass() {
	d.evictorLock.Lock()
	defer d.evictorLock.Unlock()
	for k := range nc.zonePodEvictor {
		// Function should return 'false' and a time after which it should be retried, or 'true' if it shouldn't (it succeeded).
		nc.zonePodEvictor[k].Try(func(value TimedValue) (bool, time.Duration) {
			node, err := nc.nodeLister.Get(value.Value)
			if apierrors.IsNotFound(err) {
				glog.Warningf("Node %v no longer present in nodeLister!", value.Value)
			} else if err != nil {
				glog.Warningf("Failed to get Node %v from the nodeLister: %v", value.Value, err)
			} else {
				zone := utilnode.GetZoneKey(node)
				EvictionsNumber.WithLabelValues(zone).Inc()
			}
			nodeUid, _ := value.UID.(string)
			remaining, err := deletePods(nc.kubeClient, nc.recorder, value.Value, nodeUid, nc.daemonSetStore)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("unable to evict node %q: %v", value.Value, err))
				return false, 0
			}
			if remaining {
				glog.Infof("Pods awaiting deletion due to NodeController eviction")
			}
			return true, 0
		})
	}
}

func (d *defaultPodEvictor) CancelPodEviction(node *v1.Node) error {

}

func (d *defaultPodEvictor) EvictPods(node *v1.Node, observed *v1.NodeCondition, timestamp metav1.Time) error {

}

func (d *defaultPodEvictor) AddPodEvictor(zone string) error {

}

// SwapLimiter safely swaps current limiter for this queue with the passed one if capacities or qps's differ.
func (d *defaultPodEvictor) SwapEvictorLimiter(zone string, newQPS float32) {

}