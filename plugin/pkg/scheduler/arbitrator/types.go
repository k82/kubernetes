package arbitrator

import "k8s.io/kubernetes/pkg/api/v1"

type Consumer struct {
	ID      string
	Request v1.ResourceList
}

func NewConsumer(id string) *Consumer {
	return &Consumer{
		ID: id,
		Request: make(v1.ResourceList),
	}
}

type Allocation struct {
	Consumer  *Consumer
	Deserved  v1.ResourceList
	Allocated v1.ResourceList
}

func NewAllocation(c *Consumer) *Allocation {
	return &Allocation{
		Consumer: c,
		Deserved: make(v1.ResourceList),
		Allocated: make(v1.ResourceList),
	}
}
