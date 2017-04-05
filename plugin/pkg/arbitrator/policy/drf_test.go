package policy

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/golang/glog"
	"github.com/k82cn/kube-arbitrator/pkg/util"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

func enableGlog() {
	flag.Set("alsologtostderr", "true")

	logLevel := "3"
	if ll, found := os.LookupEnv("GLOG_TEST_LEVEL"); found {
		if _, err := strconv.Atoi(ll); err == nil {
			logLevel = ll
		}
	}
	flag.Set("v", logLevel)

	flag.Parse()
}

func newConsumer(name string) *util.Consumer {
	return &util.Consumer{
		MetaData:    api.ObjectMeta{Name: name},
		Request:     util.EmptyResource(),
		PendingPods: util.NewFIFO(util.PodInfoKeyFunc),

		RunningPods: util.NewFIFO(util.PodInfoKeyFunc),
		Allocated:   util.EmptyResource(),

		Deserved: util.EmptyResource(),
	}
}

func addPod(consumer *util.Consumer, pod *util.PodInfo) {
	switch pod.Status {
	case v1.PodRunning:
		consumer.Allocated.Add(pod.Resource)
		consumer.RunningPods.Add(pod)
	case v1.PodPending:
		consumer.Request.Add(pod.Resource)
		consumer.PendingPods.Add(pod)
	default:
		glog.Warningf("Unknown Pod status for <%v/%v>", pod.Namespace, pod.Name)
	}
}

func TestDrf_Allocate(t *testing.T) {
	enableGlog()

	drf := &drf{}

	p1 := &util.PodInfo{
		Name:         "p1",
		Namespace:    "default",
		Status:       v1.PodPending,
		ConsumerName: "c1",
		Resource: &util.Resource{
			MilliCPU: 2000,
			Memory:   20 * 1024 * 1024,
		},
	}

	p2 := &util.PodInfo{
		Name:         "p1",
		Namespace:    "default",
		Status:       v1.PodRunning,
		ConsumerName: "c2",
		Resource: &util.Resource{
			MilliCPU: 2000,
			Memory:   20 * 1024 * 1024,
		},
	}

	c1 := newConsumer("c1")
	addPod(c1, p1)

	c2 := newConsumer("c2")
	addPod(c2, p2)

	n1 := &util.NodeInfo{
		Name: "n1",
		Allocatable: &util.Resource{
			MilliCPU: 8000,
			Memory:   2 * 1024 * 1024 * 1024,
		},
	}

	n2 := &util.NodeInfo{
		Name: "n2",
		Allocatable: &util.Resource{
			MilliCPU: 8000,
			Memory:   2 * 1024 * 1024 * 1024,
		},
	}

	n3 := &util.NodeInfo{
		Name: "n3",
		Allocatable: &util.Resource{
			MilliCPU: 8000,
			Memory:   2 * 1024 * 1024 * 1024,
		},
	}

	tests := []struct {
		Name      string
		Nodes     map[string]*util.NodeInfo
		Consumers map[string]*util.Consumer
	}{
		{
			Name:      "case 1",
			Nodes:     map[string]*util.NodeInfo{n1.Name: n1, n2.Name: n2, n3.Name: n3},
			Consumers: map[string]*util.Consumer{c1.MetaData.Name: c1, c2.MetaData.Name: c2},
		},
		{
			Name:      "case 2",
			Nodes:     map[string]*util.NodeInfo{n1.Name: n1, n2.Name: n2, n3.Name: n3},
			Consumers: map[string]*util.Consumer{},
		},
		{
			Name:      "case 3",
			Nodes:     map[string]*util.NodeInfo{n1.Name: n1, n2.Name: n2, n3.Name: n3},
			Consumers: map[string]*util.Consumer{c2.MetaData.Name: c2},
		},
		{
			Name:      "case 1",
			Nodes:     map[string]*util.NodeInfo{n1.Name: n1, n2.Name: n2, n3.Name: n3},
			Consumers: map[string]*util.Consumer{c1.MetaData.Name: c1},
		},
	}

	for _, test := range tests {
		t.Logf("=========== %s ===========\n", test.Name)

		drf.Allocate(test.Nodes, test.Consumers)

		for _, consumer := range test.Consumers {
			t.Logf("%s request <%v>, allocated <%v>, deserved <%v>\n", consumer.MetaData.Name, consumer.Request,
				consumer.Allocated, consumer.Deserved)
		}
		t.Log("\n")
	}
}
