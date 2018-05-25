package v1alpha1

import (
	"k8s.io/api/scheduling/v1alpha1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

func Convert_v1alpha1_PodSchedulingGroupSpec_To_scheduling_PodSchedulingGroupSpec(in *v1alpha1.PodSchedulingGroupSpec, out *scheduling.PodSchedulingGroupSpec, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_PodSchedulingGroupSpec_To_scheduling_PodSchedulingGroupSpec(in, out, s); err != nil {
		return err
	}

	return nil
}
