// Package handler dispatches informer events without an external controller library.
package handler

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Handler consumes events. Update receives the current object followed by its predecessor.
type Handler interface {
	Create(interface{}) error
	Update(interface{}, interface{}) error
	Delete(interface{}) error
}

type event struct {
	kind              string
	object, oldObject interface{}
}

// Controller preserves event order for each object while processing independent
// objects concurrently. Failed events are retried with rate limiting.
type Controller struct {
	name     string
	informer cache.SharedIndexInformer
	handler  Handler
	workers  uint
	ignored  map[string]bool
	queue    workqueue.RateLimitingInterface
	mu       sync.Mutex
	pending  map[string][]*event
}

func NewController(name string, informer cache.SharedIndexInformer, handler Handler, workers uint, ignoredNamespaces []string) *Controller {
	if workers == 0 {
		workers = 1
	}
	c := &Controller{
		name: name, informer: informer, handler: handler, workers: workers,
		ignored: make(map[string]bool), pending: make(map[string][]*event),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
	}
	for _, namespace := range ignoredNamespaces {
		c.ignored[namespace] = true
	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.enqueue(&event{kind: "create", object: obj}) },
		UpdateFunc: func(old, obj interface{}) { c.enqueue(&event{kind: "update", object: obj, oldObject: old}) },
		DeleteFunc: func(obj interface{}) {
			if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				obj = tombstone.Obj
			}
			c.enqueue(&event{kind: "delete", object: obj})
		},
	})
	return c
}

func (c *Controller) enqueue(e *event) {
	object, err := meta.Accessor(e.object)
	if err != nil {
		log.WithError(err).Error("Cannot dispatch informer object")
		return
	}
	if c.ignored[object.GetNamespace()] {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(e.object)
	if err != nil {
		log.WithError(err).Error("Cannot identify informer object")
		return
	}
	c.mu.Lock()
	c.pending[key] = append(c.pending[key], e)
	c.mu.Unlock()
	c.queue.Add(key)
}

func (c *Controller) processNext() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)
	key := item.(string)
	c.mu.Lock()
	events := c.pending[key]
	c.mu.Unlock()
	if len(events) == 0 {
		c.queue.Forget(item)
		return true
	}
	e := events[0]
	var err error
	switch e.kind {
	case "create":
		err = c.handler.Create(e.object)
	case "update":
		err = c.handler.Update(e.object, e.oldObject)
	case "delete":
		err = c.handler.Delete(e.object)
	}
	if err != nil {
		log.WithError(err).WithField("controller", c.name).Warn("Retrying informer event")
		c.queue.AddRateLimited(item)
		return true
	}
	c.queue.Forget(item)
	c.mu.Lock()
	remaining := c.pending[key][1:]
	if len(remaining) == 0 {
		delete(c.pending, key)
	} else {
		c.pending[key] = remaining
	}
	c.mu.Unlock()
	if len(remaining) != 0 {
		c.queue.Add(key)
	}
	return true
}

// Run starts the informer and stops its workers when stop is closed.
func (c *Controller) Run(stop <-chan struct{}) {
	defer c.queue.ShutDown()
	go c.informer.Run(stop)
	if !cache.WaitForCacheSync(stop, c.informer.HasSynced) {
		return
	}
	var workers sync.WaitGroup
	for i := uint(0); i < c.workers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for c.processNext() {
			}
		}()
	}
	<-stop
	c.queue.ShutDown()
	workers.Wait()
}
