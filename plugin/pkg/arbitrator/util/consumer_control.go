package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/1.5/pkg/util/json"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/cache"
)

type ThirdPartyResourceController interface {
	AddEventHandler(*cache.ResourceEventHandlerFuncs)

	Update(obj interface{}) error

	Get(tprName string) interface{}

	Run(stop <-chan interface{})
}

type consumerController struct {
	cache.ResourceEventHandlerFuncs

	mutex      sync.Mutex
	consumers  map[string]*Consumer
	restClient *rest.RESTClient
	interval   int
	namespace  string
	kind       string
}

func (cc *consumerController) AddEventHandler(hs *cache.ResourceEventHandlerFuncs) {
	cc.AddFunc = hs.AddFunc
	cc.UpdateFunc = hs.UpdateFunc
	cc.DeleteFunc = hs.DeleteFunc
}

func (cc *consumerController) Get(tprName string) interface{} {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	consumer, found := cc.consumers[tprName]
	if !found {
		return nil
	}

	res := Consumer{}

	// TODO: replace with deep copy
	raw, _ := json.Marshal(consumer)
	json.Unmarshal(raw, &res)

	return res
}

func (cc *consumerController) Update(obj interface{}) error {
	consumer, ok := obj.(*Consumer)
	if !ok {
		return fmt.Errorf("failed to convert %v to consumer type", obj)
	}

	consumer.MetaData.ResourceVersion = "0" // trigger un-condition update

	if raw, err := json.Marshal(consumer); err != nil {
		return err
	} else {
		glog.V(4).Infof("update consumer to %s", string(raw))
		result := cc.restClient.Put().
			Namespace(cc.namespace).
			Resource(cc.kind).
			Name(consumer.MetaData.Name).
			Body(raw).Do()
		if result.Error() != nil {
			return result.Error()
		}
	}

	return nil
}

func (cc *consumerController) Run(stop <-chan interface{}) {
	for {
		result := cc.restClient.Get().Namespace(cc.namespace).Resource(cc.kind).Do()

		var status int
		result.StatusCode(&status)
		if status != 200 {
			glog.Errorf("HTTP code is %d", status)
			continue
		}

		raw, err := result.Raw()
		if err != nil {
			glog.Errorf("failed to get raw result: %v", err)
			continue
		}

		consumerList, err := NewConsumerList(raw)
		if err != nil {
			glog.Errorf("failed to parse consumer: %v", err)
			continue
		}

		localCache := make(map[string]*Consumer)
		func() {
			cc.mutex.Lock()
			defer cc.mutex.Unlock()

			// Add or update by server's response
			for i := range consumerList.Items {
				consumer := &consumerList.Items[i]

				localCache[consumer.MetaData.Name] = consumer
				glog.V(4).Infof("handle consumer %s (%v)", consumer.MetaData.Name, consumer)

				c, found := cc.consumers[consumer.MetaData.Name]
				if found {
					cc.UpdateFunc(c, consumer)
				} else {
					cc.AddFunc(consumer)
				}

				cc.consumers[consumer.MetaData.Name] = consumer
			}

			// Delete consumer if not in server's response
			for _, consumer := range cc.consumers {
				if _, found := localCache[consumer.MetaData.Name]; !found {
					glog.V(4).Infof("delete consumer %s", consumer.MetaData.Name)
					cc.DeleteFunc(consumer)
					delete(cc.consumers, consumer.MetaData.Name)
				}
			}
		}()

		time.Sleep(1 * time.Second)
	}
}

func NewConsumerController(restclient *rest.RESTClient) ThirdPartyResourceController {
	return &consumerController{
		restClient: restclient,
		consumers:  map[string]*Consumer{},
		namespace:  "default",
		kind:       "consumers",
	}
}
