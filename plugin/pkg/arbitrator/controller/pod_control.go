package controller

import (
	"time"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type PodController interface {
	UnBind(pod *v1.Pod) error
	AddEventHandler(cache.ResourceEventHandlerFuncs) error

	Run(stop <-chan interface{})
}

type podController struct {
	client      *rest.RESTClient
	podInformer cache.SharedIndexInformer
}

func NewPodController(client *rest.RESTClient) PodController {
	// Watching Pending & Running Pods.
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	lw := cache.NewListWatchFromClient(client, "pods", v1.NamespaceAll, selector)

	podInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return lw.List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return lw.Watch(options)
			},
		},
		&v1.Pod{},
		1 * time.Second,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	return &podController{
		client:      client,
		podInformer: podInformer,
	}
}

func (pc *podController) Run(stop <-chan interface{}) {
	pc.podInformer.Run(nil)
}

func (pc *podController) AddEventHandler(handlers cache.ResourceEventHandlerFuncs) {
	pc.podInformer.AddEventHandler(handlers)
}

func (b *podController) UnBind(pod *v1.Pod) error {
	return b.client.Delete().Namespace(pod.Namespace).Resource("pods").Name(pod.Name).Do().Error()
}
