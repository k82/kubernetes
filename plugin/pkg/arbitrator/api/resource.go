package api

import (
	"fmt"
	"math"

	"k8s.io/client-go/pkg/api/v1"
)

type Resource struct {
	MilliCPU float64
	Memory   float64
}

func EmptyResource() *Resource {
	return &Resource{
		MilliCPU: 0,
		Memory:   0,
	}
}

func NewResource(rl v1.ResourceList) *Resource {
	cpu := rl[v1.ResourceCPU]
	mem := rl[v1.ResourceMemory]

	return &Resource{
		MilliCPU: float64(cpu.MilliValue()),
		Memory:   float64(mem.Value()),
	}
}

func CopyResource(r *Resource) *Resource {
	return &Resource{
		MilliCPU: r.MilliCPU,
		Memory:   r.Memory,
	}
}

var minMilliCPU float64 = 10
var minMemory float64 = 10 * 1024 * 1024 // 10M

func (r Resource) IsEmpty() bool {
	return r.MilliCPU < minMilliCPU && r.Memory < minMemory
}

func (r *Resource) Add(rr *Resource) *Resource {
	r.MilliCPU += rr.MilliCPU
	r.Memory += rr.Memory
	return r
}

func (r *Resource) Sub(rr *Resource) *Resource {
	r.MilliCPU -= rr.MilliCPU
	r.Memory -= rr.Memory
	return r
}

func (r *Resource) Less(rr *Resource) bool {
	return r.MilliCPU < rr.MilliCPU && r.Memory < rr.Memory
}

func (r *Resource) LessEqual(rr *Resource) bool {
	return (r.MilliCPU < rr.MilliCPU || math.Abs(r.MilliCPU-rr.MilliCPU) < 0.01) &&
		(r.Memory < rr.Memory || math.Abs(r.Memory-rr.Memory) < 1)
}

func (r Resource) String() string {
	return fmt.Sprintf("cpu %f, mem %f", r.MilliCPU, r.Memory)
}
