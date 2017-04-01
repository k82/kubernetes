package allocator

import "k8s.io/kubernetes/pkg/api/v1"

type Allocate func (req, alloc map[string]v1.ResourceList, total v1.ResourceList) map[string]v1.ResourceList