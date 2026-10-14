package handler

import (
	"errors"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

type recordingHandler struct {
	calls             []string
	fail              bool
	current, previous interface{}
}

func (h *recordingHandler) Create(obj interface{}) error {
	if h.fail {
		h.fail = false
		return errors.New("retry")
	}
	h.calls = append(h.calls, "create")
	return nil
}
func (h *recordingHandler) Update(obj, old interface{}) error {
	h.calls = append(h.calls, "update")
	h.current, h.previous = obj, old
	return nil
}
func (h *recordingHandler) Delete(obj interface{}) error {
	h.calls = append(h.calls, "delete")
	return nil
}

func newTestController(h Handler) *Controller {
	informer := cache.NewSharedIndexInformer(&cache.ListWatch{}, &unstructured.Unstructured{}, 0, nil)
	return NewController("test", informer, h, 2, []string{"ignored"})
}

func TestOrderedEventsAndRetry(t *testing.T) {
	h := &recordingHandler{fail: true}
	c := newTestController(h)
	defer c.queue.ShutDown()
	old := &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: "object", Namespace: "default"}}
	current := old.DeepCopy()
	current.ResourceVersion = "2"
	c.enqueue(&event{kind: "create", object: old})
	c.enqueue(&event{kind: "update", object: current, oldObject: old})
	c.enqueue(&event{kind: "delete", object: current})
	c.processNext() // Failed create remains ahead of update and delete.
	if len(h.calls) != 0 || c.queue.NumRequeues("default/object") != 1 {
		t.Fatal("failed event should remain queued for retry")
	}
	for i := 0; i < 3; i++ {
		c.processNext()
	}
	if !reflect.DeepEqual(h.calls, []string{"create", "update", "delete"}) {
		t.Fatalf("unexpected event order: %v", h.calls)
	}
	if h.current != current || h.previous != old {
		t.Fatal("update must receive current object before previous object")
	}
	if len(c.pending) != 0 {
		t.Fatal("completed events should be released")
	}
}

func TestIgnoredNamespaceAndMalformedObject(t *testing.T) {
	c := newTestController(&recordingHandler{})
	defer c.queue.ShutDown()
	c.enqueue(&event{kind: "create", object: &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{Name: "object", Namespace: "ignored"},
	}})
	c.enqueue(&event{kind: "create", object: "invalid"})
	if c.queue.Len() != 0 || len(c.pending) != 0 {
		t.Fatal("ignored or malformed events must not be queued")
	}
}

func TestIndependentObjectCanProgressDuringRetry(t *testing.T) {
	h := &recordingHandler{fail: true}
	c := newTestController(h)
	defer c.queue.ShutDown()
	c.enqueue(&event{kind: "create", object: &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: "first"}}})
	c.enqueue(&event{kind: "create", object: &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: "second"}}})
	c.processNext()
	c.processNext()
	if len(h.calls) != 1 {
		t.Fatalf("independent object did not progress: %v", h.calls)
	}
	if len(c.pending["first"]) != 1 {
		t.Fatal("failed first object should remain pending")
	}
}
