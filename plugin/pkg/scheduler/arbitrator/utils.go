package arbitrator

import (
	"k8s.io/kubernetes/pkg/api/v1"
)

func GetPodRequest(pod *v1.Pod) v1.ResourceList{
	result := make(v1.ResourceList)
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Requests {
			result[rName].Add(rQuantity)
		}
	}
	return result
}
