package policy

import (
	"k8s.io/kubernetes/plugin/pkg/arbitrator/api"
)

type Allocator interface {
	Allocate(nodes map[string]*api.NodeInfo, consumers map[string]*api.Consumer) map[string]*api.Allocation
}

func NewAllocator() Allocator {
	return &drf{}
}
