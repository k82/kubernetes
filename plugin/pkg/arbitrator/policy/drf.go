package policy

import (
	"fmt"
	"math"

	"github.com/golang/glog"

	"k8s.io/kubernetes/plugin/pkg/arbitrator/api"
	"k8s.io/kubernetes/plugin/pkg/arbitrator/util"
)

type consumer struct {
	name     string
	request  *api.Resource
	pods     *util.FIFO
	deserved *api.Resource
	share    float64
}

func (c consumer) String() string {
	return fmt.Sprintf("%v request <%v> and deserved <%v>", c.name, c.request, c.deserved)
}

func (c *consumer) Priority() float64 {
	return c.share
}

type drf struct {
	total     *api.Resource
	available *api.Resource
	consumers map[string]*consumer
}

func buildRequest(consumers_ map[string]*api.Consumer) map[string]*consumer {
	consumers := map[string]*consumer{}

	for _, c := range consumers_ {
		pods := util.CopyFIFO(c.RunningPods)
		pods.Append(c.PendingPods)

		consumers[c.MetaData.Name] = &consumer{
			name:     c.MetaData.Name,
			pods:     pods,
			request:  util.CopyResource(c.Request),
			deserved: util.EmptyResource(),
		}
	}

	return consumers
}

func mapToPriorityQueue(consumers map[string]*consumer) *util.PriorityQueue {
	pq := util.NewPriorityQueue()

	for _, consumer := range consumers {
		pq.Push(consumer)
	}

	return pq
}

func (d *drf) Allocate(nodes map[string]*util.NodeInfo, consumers_ map[string]*api.Consumer) {
	d.total = api.EmptyResource()
	d.available = api.EmptyResource()
	d.consumers = buildRequest(consumers_)

	if len(nodes) == 0 || len(d.consumers) == 0 {
		return
	}

	// Got allocatable resources in the cluster.
	for _, node := range nodes {
		d.total.Add(node.Allocatable)
		d.available.Add(node.Allocatable)
	}

	/*
	 * Allocate resources.
	 */
	for {
		pq := mapToPriorityQueue(d.consumers)

		allocatedOnce := false
		for {
			if d.available.IsEmpty() || pq.Len() == 0 {
				break
			}

			consumer := pq.Pop().(*consumer)
			if consumer.pods.IsEmpty() {
				continue
			}

			pod := consumer.pods.Pop().(*util.PodInfo)

			// If available resource does not have enough resource for the pod, skip it.
			if !pod.Resource.LessEqual(d.available) {
				continue
			}

			d.allocate(consumer, pod.Resource)
			consumer.share = d.calculateShare(consumer)

			allocatedOnce = true
			pq.Push(consumer)

			glog.V(4).Infof("<%s> priority is <%f> (total: <%v>, available: <%v>)",
				consumer.name, consumer.share, d.total, d.available)
		}

		if !allocatedOnce {
			break
		}
	}

	/*
	 * Update to consumer's deserved
	 */
	for _, consumer := range d.consumers {
		consumers_[consumer.name].Deserved = api.CopyResource(consumer.deserved)
	}
}

func (d *drf) deserve(consumer *consumer, request *api.Resource) {
	glog.V(4).Infof("deserve <%v> to <%v>", request, consumer)
	consumer.deserved.Add(request)
	d.available.Sub(request)
}

func (d *drf) allocate(consumer *consumer, request *api.Resource) {
	glog.V(4).Infof("allocate <%v> to <%v>", request, consumer)
	d.deserve(consumer, request)
	consumer.request.Sub(request)
}

func (d *drf) calculateShare(consumer *consumer) float64 {
	allRequest := api.EmptyResource()
	allRequest.Add(consumer.request)
	allRequest.Add(consumer.deserved)

	cpuShare := allRequest.MilliCPU / d.total.MilliCPU
	memShare := allRequest.Memory / d.total.Memory

	// if dominate resource is CPU, return its share
	if cpuShare > memShare {
		return consumer.deserved.MilliCPU / d.total.MilliCPU
	}

	// if dominate resource is memory, return its share
	if cpuShare < memShare {
		return consumer.deserved.Memory / d.total.Memory
	}

	return math.Max(consumer.deserved.MilliCPU/d.total.MilliCPU,
		consumer.deserved.Memory/d.total.Memory)
}
