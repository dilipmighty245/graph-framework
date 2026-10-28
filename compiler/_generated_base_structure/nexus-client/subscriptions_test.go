package nexus_client

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func TestSubscriptionSharesInformerAndInternalHandler(t *testing.T) {
	client := &Clientset{}
	t.Cleanup(client.UnsubscribeAll)
	var factories, adds int32
	events := watch.NewRaceFreeFake()
	factory := func() cache.SharedIndexInformer {
		atomic.AddInt32(&factories, 1)
		return cache.NewSharedIndexInformer(&cache.ListWatch{
			ListFunc: func(metav1.ListOptions) (runtime.Object, error) {
				return &unstructured.UnstructuredList{
					Object: map[string]interface{}{"metadata": map[string]interface{}{"resourceVersion": "1"}},
					Items: []unstructured.Unstructured{{Object: map[string]interface{}{
						"apiVersion": "example.test/v1", "kind": "Example",
						"metadata": map[string]interface{}{"name": "one", "resourceVersion": "1"},
					}}},
				}, nil
			},
			WatchFunc: func(metav1.ListOptions) (watch.Interface, error) { return events, nil },
		}, &unstructured.Unstructured{}, 0, cache.Indexers{})
	}
	added := make(chan struct{}, 32)
	handler := cache.ResourceEventHandlerFuncs{AddFunc: func(interface{}) {
		atomic.AddInt32(&adds, 1)
		added <- struct{}{}
	}}
	var workers sync.WaitGroup
	informers := make(chan cache.SharedIndexInformer, 32)
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			informer, err := client.subscribe("example", factory, handler)
			if err != nil {
				t.Error(err)
				return
			}
			informers <- informer
		}()
	}
	workers.Wait()
	close(informers)
	var first cache.SharedIndexInformer
	for informer := range informers {
		if first == nil {
			first = informer
		} else if first != informer {
			t.Fatal("concurrent registrations received different informers")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.WaitForCacheSync(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-added:
	case <-ctx.Done():
		t.Fatal("missing initial event")
	}
	client.UnsubscribeAll()
	if atomic.LoadInt32(&factories) != 1 || atomic.LoadInt32(&adds) != 1 {
		t.Fatalf("factories=%d internal adds=%d; want one of each", factories, adds)
	}
}

func TestSubscriptionDeletionTombstones(t *testing.T) {
	obj := &unstructured.Unstructured{}
	tombstone := cache.DeletedFinalStateUnknown{Key: "one", Obj: obj}
	for _, input := range []interface{}{obj, tombstone, &tombstone} {
		if got := subscriptionObject(input); got != obj {
			t.Fatalf("%T did not preserve the deleted object", input)
		}
	}
	var empty *cache.DeletedFinalStateUnknown
	if subscriptionObject(empty) != nil {
		t.Fatal("nil tombstone should have no object")
	}
}
