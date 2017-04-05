package util

import (
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

const (
	SCHED_MGR_NAME = "kube-arbitrator"
	KSM_GROUP      = "kube-gantt.k82.me"
	KSM_VERSION    = "v1"
	KSM_API_PATH   = "/apis"
)

type Builder struct {
	Recorder     record.EventRecorder
	RESTClient   *rest.RESTClient
	KubeClient   *kubernetes.Clientset
	ResyncPeriod time.Duration
}

func NewBuilder(kubeConfig *rest.Config) (*Builder, error) {
	builder := &Builder{}
	rest.AddUserAgent(kubeConfig, SCHED_MGR_NAME)

	// Build KubeClient
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	builder.KubeClient = kubeClient

	// Build Recorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	//eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.Core().Events("")})
	builder.Recorder = eventBroadcaster.NewRecorder(v1.EventSource{Component: SCHED_MGR_NAME})

	// Build restful client.
	kubeConfig.GroupVersion = &unversioned.GroupVersion{
		Group:   KSM_GROUP,
		Version: KSM_VERSION,
	}

	kubeConfig.NegotiatedSerializer = api.Codecs
	kubeConfig.APIPath = KSM_API_PATH

	client, err := rest.RESTClientFor(kubeConfig)
	if err != nil {
		return nil, err
	}
	builder.RESTClient = client

	// Return
	return builder, nil
}

func (b *Builder) ConsumerController() ThirdPartyResourceController {
	return NewConsumerController(b.RESTClient)
}

func (b *Builder) NodeInformer() cache.SharedIndexInformer {
	sharedIndexInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return b.KubeClient.Core().Nodes().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return b.KubeClient.Core().Nodes().Watch(options)
			},
		},
		&v1.Node{},
		b.ResyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	return sharedIndexInformer
}

func (b *Builder) PodController() PodController {
	return NewPodController(b.RESTClient)
}
