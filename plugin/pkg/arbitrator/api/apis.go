package api

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/client-go/pkg/api"
	"k8s.io/apimachinery/pkg/util/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Consumer struct {
	metav1.TypeMeta
	MetaData  api.ObjectMeta `json:"metadata"`

	Request   *Resource `json:"request"`

	Pending   *Resource `json:"pending"`
	Allocated *Resource `json:"allocated"`

	Share     float64 `json:"-"`
}

type Allocation struct {
	metav1.TypeMeta
	MetaData api.ObjectMeta `json:"metadata"`

	Consumer string `json:"consumer"`

	Deserved *Resource `json:"deserved"`
}

func (o Consumer) String() string {
	return fmt.Sprintf("%v request <%v>", o.MetaData.Name, o.Request)
}

func (o *Consumer) Priority() float64 {
	return o.Share
}

/**
 * Definitation of ConsumerList
 */
type ConsumerList struct {
	metav1.TypeMeta
	MetaData api.ObjectMeta

	Items    []Consumer
}

func NewConsumerList(raw []byte) (*ConsumerList, error) {
	glog.V(6).Infof("construct ConsumerList from %s", string(raw))

	consumerList := &ConsumerList{}

	if err := json.Unmarshal(raw, consumerList); err != nil {
		return nil, err
	}

	return consumerList, nil
}

// ---------> End of ConsumerList

type NodeInfo struct {

}