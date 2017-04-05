package arbitrator

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/k82cn/kube-arbitrator/pkg/policy"
	"github.com/k82cn/kube-arbitrator/pkg/util"

	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/tools/cache"
	"k8s.io/client-go/1.5/tools/record"
)

type SchedulerArbitrator interface {
	Allocatable(pod *util.PodInfo) bool

	Run()
}

type ReclaimRequest struct {
	name string
	time time.Time
}

func ReclaimRequestKeyFunc(obj interface{}) (string, error) {
	if rr, ok := obj.(*ReclaimRequest); ok {
		return rr.name, nil
	}

	return "", fmt.Errorf("failed to conver %v to *ReclaimRequest.", obj)
}

type schedulerArbitrator struct {
	mutex sync.Mutex

	nodes     map[string]*util.NodeInfo // key: hostname, value: *v1.Node
	consumers map[string]*util.Consumer

	reclaimRequest  *util.FIFO
	terminating     map[string]*util.Resource
	terminatingPods *util.FIFO

	recorder  record.EventRecorder
	allocator policy.Allocator

	consumerControl util.ThirdPartyResourceController
	nodeInformer    cache.SharedIndexInformer
	podControl      util.PodController
}

func NewSchedulerArbitrator(consumerControl util.ThirdPartyResourceController,
	podControl util.PodController,
	nodeInformer cache.SharedIndexInformer,
	recorder record.EventRecorder,
) SchedulerArbitrator {
	sa := &schedulerArbitrator{
		nodes:           make(map[string]*util.NodeInfo),
		consumers:       make(map[string]*util.Consumer),
		reclaimRequest:  util.NewFIFO(ReclaimRequestKeyFunc),
		terminating:     make(map[string]*util.Resource),
		terminatingPods: util.NewFIFO(util.PodInfoKeyFunc),
		recorder:        recorder,
	}

	consumerControl.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sa.AddConsumer(obj)
		},
		DeleteFunc: func(obj interface{}) {
			sa.DeleteConsumer(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			sa.UpdateConsumer(obj)
		},
	})

	podControl.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sa.AddPod(obj)
		},
		DeleteFunc: func(obj interface{}) {
			sa.DeletePod(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			sa.UpdatePod(old, obj)
		},
	})

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sa.AddNode(obj)
		},
		DeleteFunc: func(obj interface{}) {
			sa.DeleteNode(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			sa.UpdateNode(obj)
		},
	})

	sa.allocator = policy.NewAllocator()
	sa.consumerControl = consumerControl
	sa.nodeInformer = nodeInformer
	sa.podControl = podControl

	return sa
}

func (sa *schedulerArbitrator) AddPod(pod__ interface{}) {
	glog.V(7).Infof("added pod <%v>", pod__)

	pod_, ok := pod__.(*v1.Pod)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Pod", pod__)
		return
	}

	pod := util.NewPodInfo(pod_)

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	consumer, found := sa.consumers[pod.ConsumerName]
	if !found {
		glog.Warningf("can not find consumer <%v>, ignore pod <%v>", pod.ConsumerName, pod)
		return
	}

	switch pod.Status {
	case v1.PodRunning:
		consumer.Allocated.Add(pod.Resource)
		consumer.RunningPods.Add(pod)
	case v1.PodPending:
		consumer.Request.Add(pod.Resource)
		consumer.PendingPods.Add(pod)
	default:
		glog.Warningf("Unknown Pod status for <%v/%v>", pod.Namespace, pod.Name)
	}
}

func (sa *schedulerArbitrator) UpdatePod(oldPod__, pod__ interface{}) {
	pod_, ok := pod__.(*v1.Pod)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Pod", pod__)
		return
	}

	oldPod_, ok := oldPod__.(*v1.Pod)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Pod", oldPod__)
		return
	}

	pod := util.NewPodInfo(pod_)
	oldPod := util.NewPodInfo(oldPod_)

	if pod.ConsumerName != oldPod.ConsumerName {
		glog.Errorf("update pod with different scheduler name: old <%v>, new <%v>", oldPod, pod)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	consumer, found := sa.consumers[pod.ConsumerName]
	if !found {
		glog.Warningf("can not find consumer <%v>, ignore pod <%v>", pod.ConsumerName, pod)
		return
	}

	// Remove resource by old pod.
	switch oldPod.Status {
	case v1.PodRunning:
		consumer.Allocated.Sub(oldPod.Resource)
		consumer.RunningPods.Delete(pod)
	case v1.PodPending:
		consumer.Request.Sub(oldPod.Resource)
		consumer.PendingPods.Delete(oldPod)
	default:
		glog.Warningf("Unknown old Pod status for <%v/%v>", oldPod.Namespace, oldPod.Name)
	}

	// Add resource for new pod.
	switch pod.Status {
	case v1.PodRunning:
		consumer.Allocated.Add(pod.Resource)
		consumer.RunningPods.Add(pod)
	case v1.PodPending:
		consumer.Request.Add(pod.Resource)
		consumer.PendingPods.Add(pod)
	default:
		glog.Warningf("Unknown Pod status for <%v/%v>", pod.Namespace, pod.Name)
	}
}

func (sa *schedulerArbitrator) DeletePod(pod__ interface{}) {
	glog.V(7).Infof("delete pod <%v>", pod__)

	pod_, ok := pod__.(*v1.Pod)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Pod", pod__)
		return
	}

	pod := util.NewPodInfo(pod_)

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	if consumer, found := sa.consumers[pod.ConsumerName]; found {
		switch pod.Status {
		case v1.PodRunning:
			consumer.Allocated.Sub(pod.Resource)
			consumer.RunningPods.Delete(pod)
		case v1.PodPending:
			consumer.Request.Sub(pod.Resource)
			consumer.PendingPods.Delete(pod)
		default:
			glog.Warningf("Unknown Pod status for <%v/%v>", pod.Namespace, pod.Name)
		}
	}

	// If it's terminating pods, update counters.
	if sa.terminatingPods.Contain(pod) {
		sa.terminatingPods.Delete(pod)
		sa.terminating[pod.ConsumerName].Sub(pod.Resource)
	}
}

func (sa *schedulerArbitrator) AddNode(node_ interface{}) {
	glog.V(7).Infof("added node <%v>", node_)

	node, ok := node_.(*v1.Node)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Node", node_)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	sa.nodes[node.Name] = util.NewNodeInfo(node)
}

func (sa *schedulerArbitrator) DeleteNode(node_ interface{}) {
	glog.V(7).Infof("remove node <%v>", node_)

	node, ok := node_.(*v1.Node)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Node", node_)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	delete(sa.nodes, node.Name)
}

func (sa *schedulerArbitrator) UpdateNode(node_ interface{}) {
	glog.V(7).Infof("update node <%v>", node_)

	node, ok := node_.(*v1.Node)
	if !ok {
		glog.Errorf("failed to convert %v to *v1.Node", node_)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	sa.nodes[node.Name] = util.NewNodeInfo(node)
}

func (sa *schedulerArbitrator) AddConsumer(consumer_ interface{}) {
	consumer, ok := consumer_.(*util.Consumer)

	if !ok {
		glog.Errorf("failed to conver %v to *util.Consumer", consumer_)
		return
	}

	consumer.Request = util.EmptyResource()
	consumer.Allocated = util.EmptyResource()
	consumer.PendingPods = util.NewFIFO(util.PodInfoKeyFunc)
	consumer.RunningPods = util.NewFIFO(util.PodInfoKeyFunc)

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	sa.consumers[consumer.MetaData.Name] = consumer
}

func (sa *schedulerArbitrator) DeleteConsumer(consumer_ interface{}) {
	glog.V(3).Infof("remove consumer <%v>", consumer_)
	consumer, ok := consumer_.(*util.Consumer)
	if !ok {
		glog.Errorf("failed to conver %v to *util.Consumer", consumer_)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	_, found := sa.consumers[consumer.MetaData.Name]
	if !found {
		glog.Errorf("Can not found consumer %s when removing consumer cache", consumer.MetaData.Name)
		return
	}

	delete(sa.consumers, consumer.MetaData.Name)
}

func (sa *schedulerArbitrator) UpdateConsumer(consumer_ interface{}) {
	consumer, ok := consumer_.(*util.Consumer)
	if !ok {
		glog.Errorf("failed to conver %v to *util.Consumer", consumer_)
		return
	}

	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	old, found := sa.consumers[consumer.MetaData.Name]
	if !found {
		glog.Errorf("Can not found consumer %s when updating consumer cache", consumer.MetaData.Name)
		return
	}

	// Arbitrator own the latest version of the following members.
	consumer.Deserved = old.Deserved

	consumer.Allocated = old.Allocated
	consumer.RunningPods = old.RunningPods

	consumer.Request = old.Request
	consumer.PendingPods = old.PendingPods

	sa.consumers[consumer.MetaData.Name] = consumer
}

func (sa *schedulerArbitrator) Allocatable(pod *util.PodInfo) bool {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	consumer, exist := sa.consumers[pod.ConsumerName]
	if !exist {
		glog.V(3).Infof("failed to find consumer for pod <%v/%v> by <%v>", pod.Namespace, pod.Name, pod.ConsumerName)
		return false
	}

	allocated := consumer.Allocated
	deserved := consumer.Deserved

	if allocated == nil || deserved == nil {
		glog.V(3).Infof("waiting for arbitrator's allocation")
		return false
	}

	if pod.Resource.Memory+allocated.Memory > deserved.Memory {
		glog.V(3).Infof("failed to allocate mem <%f> to pod <%v/%v> (allocated <%f>, deserved <%f>)",
			pod.Resource.Memory,
			pod.Namespace,
			pod.Name,
			allocated.Memory,
			deserved.Memory,
		)
		return false
	}

	if pod.Resource.MilliCPU+allocated.MilliCPU > deserved.MilliCPU {
		glog.V(3).Infof("failed to allocate cpu <%f> to pod <%v/%v> (allocated <%f>, deserved <%f>)",
			pod.Resource.MilliCPU,
			pod.Namespace,
			pod.Name,
			allocated.MilliCPU,
			deserved.MilliCPU,
		)
		return false
	}

	return true
}

func (sa *schedulerArbitrator) isReclaimPeriodExpired(t time.Time) bool {
	return time.Now().After(t.Add(5 * time.Second))
}

func (sa *schedulerArbitrator) reclaim() {
	for {
		reclaimRequests := sa.reclaimRequest.List()

		for _, reclaimRequest_ := range reclaimRequests {
			reclaimRequest := reclaimRequest_.(*ReclaimRequest)

			// if recliamPeriod not expired, skip it.
			if !sa.isReclaimPeriodExpired(reclaimRequest.time) {
				continue
			}

			func() {
				sa.mutex.Lock()
				defer sa.mutex.Unlock()

				consumer := sa.consumers[reclaimRequest.name]
				reclaimRes := sa.ReclaimRequest(reclaimRequest.name)

				for _, pod_ := range consumer.RunningPods.List() {
					pod := pod_.(*util.PodInfo)

					if reclaimRes.IsEmpty() {
						break
					}

					// If reclaim successfully, update counters.
					if err := sa.TermintePod(pod); err == nil {
						reclaimRes.Sub(pod.Resource)
					}
				}
			}()
		}

		time.Sleep(1 * time.Second)
	}
}

func (sa *schedulerArbitrator) TermintePod(pod *util.PodInfo) error {
	// If reclaim successfully, update counters.
	if err := sa.podControl.UnBind(pod); err != nil {
		return err
	}

	sa.terminating[pod.ConsumerName].Add(pod.Resource)
	sa.terminatingPods.Add(pod)
	return nil
}

func (sa *schedulerArbitrator) ReclaimRequest(name string) *util.Resource {
	consumer := sa.consumers[name]
	terminating, found := sa.terminating[name]
	if !found {
		sa.terminating[name] = util.EmptyResource()
		return util.CopyResource(consumer.Allocated).Sub(consumer.Deserved)
	}
	return util.CopyResource(consumer.Allocated).Sub(consumer.Deserved).Sub(terminating)
}

func (sa *schedulerArbitrator) Run() {
	// start reclaim go routine
	go sa.reclaim()

	// start to allocate resources to consumers.
	for {
		func() {
			sa.mutex.Lock()
			defer sa.mutex.Unlock()

			sa.allocator.Allocate(sa.nodes, sa.consumers)

			for _, consumer := range sa.consumers {
				// If overused, send reclaim request to recliam go routine.
				reclaimRes := sa.ReclaimRequest(consumer.MetaData.Name)
				if !reclaimRes.IsEmpty() {
					sa.reclaimRequest.AddIfNotPresent(&ReclaimRequest{
						name: consumer.MetaData.Name,
						time: time.Now(),
					})
				}

				if err := sa.consumerControl.Update(consumer); err != nil {
					glog.Errorf("failed to update consumer <%s>: %v", consumer.MetaData.Name, err)
				}
			}
		}()

		time.Sleep(1 * time.Second)
	}
}
