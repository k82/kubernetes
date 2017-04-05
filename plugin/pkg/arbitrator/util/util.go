package util

import (
	"fmt"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/arbitrator/api"
)

type PodInfo struct {
	Name         string
	Namespace    string
	ConsumerName string
	Status       v1.PodPhase
	Hostname     string
	Resource     *api.Resource
}

func NewPodInfo(pod *v1.Pod) *PodInfo {
	return &PodInfo{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		ConsumerName: pod.Namespace,
		Status:       pod.Status.Phase,
		Hostname:     pod.Spec.NodeName,
		Resource:     getResourceRequest(pod),
	}
}

func (p PodInfo) String() string {
	return fmt.Sprintf("%v/%v", p.Namespace, p.Name)
}

func PodInfoKeyFunc(obj interface{}) (string, error) {
	if pod, ok := obj.(*PodInfo); ok {
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), nil
	}

	return "", fmt.Errorf("failed to convert <%v> to *util.PodInfo", obj)
}

func NodeInfoKeyFunc(obj interface{}) (string, error) {
	if node, ok := obj.(*NodeInfo); ok {
		return fmt.Sprintf("%s", node.Name), nil
	}

	return "", fmt.Errorf("failed to convert <%v> to *util.NodeInfo", obj)
}

type NodeInfo struct {
	Name             string
	Allocatable      *Resource
	RequestResources *Resource
	Capacity         *Resource
}

func NewNodeInfo(node *v1.Node) *NodeInfo {
	return &NodeInfo{
		Name:             node.Name,
		Allocatable:      NewResource(node.Status.Allocatable),
		Capacity:         NewResource(node.Status.Capacity),
		RequestResources: EmptyResource(),
	}
}

func (ni *NodeInfo) AddPod(pi *PodInfo) {
	if ni == nil || pi == nil {
		return
	}
	ni.RequestResources.Add(pi.Resource)
}

func (ni *NodeInfo) DeletePod(pi *PodInfo) {
	if ni == nil || pi == nil {
		return
	}
	ni.RequestResources.Sub(pi.Resource)
}

func getResourceRequest(pod *v1.Pod) *Resource {
	result := Resource{}
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Requests {
			switch rName {
			case v1.ResourceMemory:
				result.Memory += float64(rQuantity.Value())
			case v1.ResourceCPU:
				result.MilliCPU += float64(rQuantity.MilliValue())
			default:
				continue
			}
		}
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		for rName, rQuantity := range container.Resources.Requests {
			switch rName {
			case v1.ResourceMemory:
				if mem := float64(rQuantity.Value()); mem > result.Memory {
					result.Memory = mem
				}
			case v1.ResourceCPU:
				if cpu := float64(rQuantity.MilliValue()); cpu > result.MilliCPU {
					result.MilliCPU = cpu
				}
			default:
				continue
			}
		}
	}
	return &result
}
