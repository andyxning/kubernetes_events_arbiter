package kubernetes

import (
	"fmt"
	"github.com/andyxning/eventarbiter/models"
	"github.com/golang/glog"
	backend "k8s.io/heapster/events/sources/kubernetes"
	"k8s.io/kubernetes/pkg/api"
	"net/url"
	"time"
)

const (
	fetchInterval = 500 * time.Millisecond
)

type kubernetes struct {
	fetchTicker *time.Ticker
	upstream    *backend.KubernetesEventSource
}

func MustNewKubernetes(uri *url.URL) models.Source {
	upstream, err := backend.NewKubernetesSource(uri)
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}

	return kubernetes{
		fetchTicker: time.NewTicker(fetchInterval),
		upstream:    upstream,
	}
}

func (k8s kubernetes) Start(eventChan chan<- *api.Event) {
	go func() {
		for {
			select {
			case <-k8s.fetchTicker.C:
				eventBatch := k8s.upstream.GetNewEvents()
				glog.Infof("got %d new events at %s", len(eventBatch.Events), eventBatch.Timestamp)

				for _, event := range eventBatch.Events {
					select {
					case eventChan <- event:
						glog.V(3).Infof("%#v", event)
					default:
						glog.Errorf("event channel is full. ignoring %#v", event)
					}
				}
			}
		}
	}()
}

func (k8s kubernetes) Stop() {
	// Nothing to do now.
}
