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

package schedulinggroup

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	scheduling "k8s.io/api/scheduling/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	schedinformers "k8s.io/client-go/informers/scheduling/v1alpha1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	schedlisters "k8s.io/client-go/listers/scheduling/v1alpha1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/kubernetes/pkg/controller"
)

const (
	statusUpdateRetries = 2

	SCHEDULING_GROUP_STR = "schedulinggroup"
)

type SchedulingGroupController struct {
	kubeClient clientset.Interface

	psgLister       schedlisters.PodSchedulingGroupLister
	psgListerSynced cache.InformerSynced

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	// PodSchedulingGroup keys that need to be synced.
	queue        workqueue.RateLimitingInterface
	recheckQueue workqueue.DelayingInterface

	broadcaster record.EventBroadcaster
	recorder    record.EventRecorder
}

func NewSchedulingGroupController(
	podInformer coreinformers.PodInformer,
	psgInformer schedinformers.PodSchedulingGroupInformer,
	kubeClient clientset.Interface,
) *SchedulingGroupController {
	dc := &SchedulingGroupController{
		kubeClient:   kubeClient,
		queue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), SCHEDULING_GROUP_STR),
		recheckQueue: workqueue.NewNamedDelayingQueue("schedulinggroup-recheck"),
		broadcaster:  record.NewBroadcaster(),
	}
	dc.recorder = dc.broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "controllermanager"})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dc.addPod,
		UpdateFunc: dc.updatePod,
		DeleteFunc: dc.deletePod,
	})
	dc.podLister = podInformer.Lister()
	dc.podListerSynced = podInformer.Informer().HasSynced

	psgInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    dc.addPSG,
			UpdateFunc: dc.updatePSG,
			DeleteFunc: dc.removePSG,
		},
		30*time.Second,
	)
	dc.psgLister = psgInformer.Lister()
	dc.psgListerSynced = psgInformer.Informer().HasSynced

	return dc
}

func (dc *SchedulingGroupController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer dc.queue.ShutDown()

	glog.Infof("Starting schedulinggroup controller")
	defer glog.Infof("Shutting down schedulinggroup controller")

	if !controller.WaitForCacheSync(SCHEDULING_GROUP_STR, stopCh, dc.podListerSynced, dc.psgListerSynced) {
		return
	}

	if dc.kubeClient != nil {
		glog.Infof("Sending events to api server.")
		dc.broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: dc.kubeClient.CoreV1().Events("")})
	} else {
		glog.Infof("No api server defined - no events will be sent to API server.")
	}

	go wait.Until(dc.worker, time.Second, stopCh)
	go wait.Until(dc.recheckWorker, time.Second, stopCh)

	<-stopCh
}

func (dc *SchedulingGroupController) addPSG(obj interface{}) {
	psg := obj.(*scheduling.PodSchedulingGroup)
	glog.V(4).Infof("add PSG %q", psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) updatePSG(old, cur interface{}) {
	psg := cur.(*scheduling.PodSchedulingGroup)
	glog.V(4).Infof("update PSG %q", psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) removePSG(obj interface{}) {
	psg := obj.(*scheduling.PodSchedulingGroup)
	glog.V(4).Infof("remove PSG %q", psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	glog.V(4).Infof("addPod called on pod %q", pod.Name)
	psg := dc.getPSGForPod(pod)
	if psg == nil {
		glog.V(4).Infof("No matching PSG for pod %q", pod.Name)
		return
	}
	glog.V(4).Infof("addPod %q -> PSG %q", pod.Name, psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) updatePod(old, cur interface{}) {
	pod := cur.(*v1.Pod)
	glog.V(4).Infof("updatePod called on pod %q", pod.Name)
	psg := dc.getPSGForPod(pod)
	if psg == nil {
		glog.V(4).Infof("No matching psg for pod %q", pod.Name)
		return
	}
	glog.V(4).Infof("updatePod %q -> PSG %q", pod.Name, psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the pod
	// changed labels the new ReplicaSet will not be woken up till the periodic
	// resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			glog.Errorf("Couldn't get object from tombstone %+v", obj)
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			glog.Errorf("Tombstone contained object that is not a pod %+v", obj)
			return
		}
	}
	glog.V(4).Infof("deletePod called on pod %q", pod.Name)
	psg := dc.getPSGForPod(pod)
	if psg == nil {
		glog.V(4).Infof("No matching PSG for pod %q", pod.Name)
		return
	}
	glog.V(4).Infof("deletePod %q -> PSG %q", pod.Name, psg.Name)
	dc.enqueuePSG(psg)
}

func (dc *SchedulingGroupController) enqueuePSG(psg *scheduling.PodSchedulingGroup) {
	key, err := controller.KeyFunc(psg)
	if err != nil {
		glog.Errorf("Cound't get key for PodSchedulingGroup object %+v: %v", psg, err)
		return
	}
	dc.queue.Add(key)
}

func (dc *SchedulingGroupController) enqueuePSGForRecheck(psg *scheduling.PodSchedulingGroup, delay time.Duration) {
	key, err := controller.KeyFunc(psg)
	if err != nil {
		glog.Errorf("Cound't get key for PodSchedulingGroup object %+v: %v", psg, err)
		return
	}
	dc.recheckQueue.AddAfter(key, delay)
}

// getPodSchedulingGroups returns a list of PodSchedulingGroup matching a pod.
// Returns an error only if no matching PodSchedulingGroup are found.
func (dc *SchedulingGroupController) getPodSchedulingGroups(pod *v1.Pod) ([]*scheduling.PodSchedulingGroup, error) {
	var selector labels.Selector

	if len(pod.Labels) == 0 {
		return nil, fmt.Errorf("no PodSchedulingGroup found for pod %v because it has no labels", pod.Name)
	}

	list, err := dc.psgLister.PodSchedulingGroups(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var psgList []*scheduling.PodSchedulingGroup
	for i := range list {
		psg := list[i]
		selector, err = metav1.LabelSelectorAsSelector(psg.Spec.Selector)
		if err != nil {
			glog.Warningf("invalid selector: %v", err)
			continue
		}

		// If a PSG with a nil or empty selector creeps in, it should match nothing, not everything.
		if selector.Empty() || !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		psgList = append(psgList, psg)
	}

	if len(psgList) == 0 {
		return nil, fmt.Errorf("could not find PodSchedulingGroup for pod %s in namespace %s with labels: %v", pod.Name, pod.Namespace, pod.Labels)
	}

	return psgList, nil
}

func (dc *SchedulingGroupController) getPSGForPod(pod *v1.Pod) *scheduling.PodSchedulingGroup {
	// getPodSchedulingGroups returns an error only if no
	// PodSchedulingGroup are found.  We don't return that as an error to the
	// caller.
	psgs, err := dc.getPodSchedulingGroups(pod)
	if err != nil {
		glog.V(4).Infof("No PodSchedulingGroup found for pod %v, PodSchedulingGroup controller will avoid syncing.", pod.Name)
		return nil
	}

	if len(psgs) > 1 {
		msg := fmt.Sprintf("Pod %q/%q matches multiple PodSchedulingGroup.  Chose %q arbitrarily.",
			pod.Namespace, pod.Name, psgs[0].Name)
		glog.Warning(msg)
		dc.recorder.Event(pod, v1.EventTypeWarning, "MultiplePodSchedulingGroups", msg)
	}
	return psgs[0]
}

// This function returns pods using the PodSchedulingGroup object.
// IMPORTANT NOTE : the returned pods should NOT be modified.
func (dc *SchedulingGroupController) getPodsForPsg(psg *scheduling.PodSchedulingGroup) ([]*v1.Pod, error) {
	sel, err := metav1.LabelSelectorAsSelector(psg.Spec.Selector)
	if sel.Empty() {
		return []*v1.Pod{}, nil
	}
	if err != nil {
		return []*v1.Pod{}, err
	}
	pods, err := dc.podLister.Pods(psg.Namespace).List(sel)
	if err != nil {
		return []*v1.Pod{}, err
	}
	return pods, nil
}

func (dc *SchedulingGroupController) worker() {
	for dc.processNextWorkItem() {
	}
}

func (dc *SchedulingGroupController) processNextWorkItem() bool {
	dKey, quit := dc.queue.Get()
	if quit {
		return false
	}
	defer dc.queue.Done(dKey)

	err := dc.sync(dKey.(string))
	if err == nil {
		dc.queue.Forget(dKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("Error syncing PodSchedulingGroup %v, requeuing: %v", dKey.(string), err))
	dc.queue.AddRateLimited(dKey)

	return true
}

func (dc *SchedulingGroupController) recheckWorker() {
	for dc.processNextRecheckWorkItem() {
	}
}

func (dc *SchedulingGroupController) processNextRecheckWorkItem() bool {
	dKey, quit := dc.recheckQueue.Get()
	if quit {
		return false
	}
	defer dc.recheckQueue.Done(dKey)
	dc.queue.AddRateLimited(dKey)
	return true
}

func (dc *SchedulingGroupController) sync(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing PodSchedulingGroup %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	psg, err := dc.psgLister.PodSchedulingGroups(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(4).Infof("PodSchedulingGroup %q has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	if err := dc.trySync(psg); err != nil {
		glog.Errorf("Failed to sync PSG %s/%s: %v", psg.Namespace, psg.Name, err)
	}

	return nil
}

func (dc *SchedulingGroupController) trySync(psg *scheduling.PodSchedulingGroup) error {
	pods, err := dc.getPodsForPsg(psg)
	if err != nil {
		dc.recorder.Eventf(psg, v1.EventTypeWarning, "NoPods", "Failed to get pods: %v", err)
		return err
	}
	if len(pods) == 0 {
		return nil
	}

	status := scheduling.PodSchedulingGroupStatus{
		MinAvailable: *(psg.Spec.MinAvailable),
		Replicas:     int32(len(pods)),
	}

	for _, pod := range pods {
		switch pod.Status.Phase {
		case v1.PodPending:
			status.Pending++
		case v1.PodRunning:
			status.Running++
		case v1.PodFailed:
			status.Failed++
		case v1.PodSucceeded:
			status.Succeeded++
		}
	}

	psgClient := dc.kubeClient.SchedulingV1alpha1().PodSchedulingGroups(psg.Namespace)

	for i := 0; i < statusUpdateRetries; i++ {
		newPsg, err := psgClient.Get(psg.Name, metav1.GetOptions{})
		if err != nil {
			newPsg = psg
		}

		newPsg.Status = status
		if _, err = psgClient.UpdateStatus(newPsg); err != nil {
			glog.Warningf("Failed to update PSG %v/%v status %v: %v",
				psg.Namespace, psg.Name, status, err)
			continue
		}
	}

	if err != nil {
		dc.enqueuePSGForRecheck(psg, 10*time.Second)
	}

	return err
}
