//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/config.tsm.tanzu.vmware.com/v1"
	gnsv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/gns.tsm.tanzu.vmware.com/v1"
	rootv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/root.tsm.tanzu.vmware.com/v1"
	nexusclient "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/nexus-client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type nodeEvent struct{ kind, name, before, after string }

func TestGeneratedClientSubscriptions(t *testing.T) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		t.Fatal("KUBECONFIG must identify the disposable framework test cluster")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	config.QPS, config.Burst = 100, 200
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	newClient := func() *nexusclient.Clientset {
		client, err := nexusclient.NewForConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(client.UnsubscribeAll)
		return client
	}
	writer, first, second := newClient(), newClient(), newClient()
	firstEvents, secondEvents := make(chan nodeEvent, 128), make(chan nodeEvent, 128)
	var dropped atomic.Int64
	value := func(obj *nexusclient.ConfigConfig) string {
		if obj.Spec.MyStr0 == nil {
			return ""
		}
		return string(*obj.Spec.MyStr0)
	}
	register := func(client *nexusclient.Clientset, events chan nodeEvent) {
		emit := func(event nodeEvent) {
			select {
			case events <- event:
			default:
				dropped.Add(1)
			}
		}
		registration, err := client.RootRoot().Config("default").RegisterEventHandler(
			func(obj *nexusclient.ConfigConfig) { emit(nodeEvent{kind: "add", name: obj.Name, after: value(obj)}) },
			func(old, current *nexusclient.ConfigConfig) {
				emit(nodeEvent{kind: "update", name: current.Name, before: value(old), after: value(current)})
			},
			func(obj *nexusclient.ConfigConfig) {
				emit(nodeEvent{kind: "delete", name: obj.Name, before: value(obj)})
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if registration == nil {
			t.Fatal("missing event registration")
		}
	}
	waitFor := func(description string, condition func() bool) {
		t.Helper()
		deadline := time.NewTimer(20 * time.Second)
		defer deadline.Stop()
		tick := time.NewTicker(20 * time.Millisecond)
		defer tick.Stop()
		for !condition() {
			select {
			case <-deadline.C:
				t.Fatalf("timed out waiting for %s", description)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-tick.C:
			}
		}
	}
	awaitEvent := func(events <-chan nodeEvent, expected nodeEvent) {
		t.Helper()
		deadline := time.NewTimer(20 * time.Second)
		defer deadline.Stop()
		for {
			select {
			case got := <-events:
				if got == expected {
					return
				}
				t.Logf("waiting for %+v; received %+v", expected, got)
			case <-deadline.C:
				t.Fatalf("missing subscription event: %+v", expected)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
	}
	register(first, firstEvents)
	if second.RootRoot().Config("default").IsSubscribed() {
		t.Fatal("a new client inherited another client's subscription")
	}
	second.SubscribeAll()
	register(second, secondEvents)

	t.Log("PASS: independent clients and event handler registration")

	root, err := writer.AddRootRoot(ctx, &rootv1.Root{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := writer.DeleteRootRoot(context.Background()); err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("root cleanup: %v", err)
		}
	})
	before := gnsv1.MyStr("before")
	node, err := root.AddConfig(ctx, &configv1.Config{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
		Spec: configv1.ConfigSpec{
			MyStr0: &before, MyStr1: []gnsv1.MyStr{}, MyStr2: map[string]gnsv1.MyStr{},
			XYZPort: 8080, ABCHost: []gnsv1.Host{}, ClusterNamespaces: []configv1.ClusterNamespace{},
			TestValMarkers: configv1.TestValMarkers{MyStr: "ab", MyInt: 3, MySlice: []string{"a", "b"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	configDeleted := false
	t.Cleanup(func() {
		if configDeleted {
			return
		}
		if err := root.DeleteConfig(context.Background()); err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("config cleanup: %v", err)
		}
	})
	name := node.Name
	for _, events := range []<-chan nodeEvent{firstEvents, secondEvents} {
		awaitEvent(events, nodeEvent{kind: "add", name: name, after: "before"})
	}
	for _, reader := range []*nexusclient.Clientset{first, second} {
		waitFor("parent-child cache", func() bool {
			child, err := reader.RootRoot().GetConfig(ctx, "default")
			return err == nil && child.Name == name && child.Labels["roots.root.tsm.tanzu.vmware.com"] == "default"
		})
	}
	t.Log("PASS: SDK node creation, add callbacks, hashed names, and hierarchy caches")

	cached, err := first.Config().GetConfigByName(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	changed := gnsv1.MyStr("local-only")
	cached.Spec.MyStr0 = &changed
	again, err := first.Config().GetConfigByName(ctx, name)
	if err != nil || value(again) != "before" {
		t.Fatalf("caller mutation changed the informer cache: %v", err)
	}
	t.Log("PASS: cached read objects are safe for callers to modify")

	update := func(text string) {
		next := gnsv1.MyStr(text)
		node.Spec.MyStr0 = &next
		if err := node.Update(ctx); err != nil {
			t.Fatal(err)
		}
	}
	update("after")
	for _, events := range []<-chan nodeEvent{firstEvents, secondEvents} {
		awaitEvent(events, nodeEvent{kind: "update", name: name, before: "before", after: "after"})
	}
	t.Log("PASS: SDK updates and old/new values in both update callbacks")

	first.RootRoot().Config("default").Unsubscribe()
	if first.RootRoot().Config("default").IsSubscribed() || !second.RootRoot().Config("default").IsSubscribed() {
		t.Fatal("unsubscribe affected the wrong client")
	}
	update("after-unsubscribe")
	awaitEvent(secondEvents, nodeEvent{kind: "update", name: name, before: "after", after: "after-unsubscribe"})
	select {
	case event := <-firstEvents:
		t.Fatalf("event delivered after unsubscribe: %+v", event)
	default:
	}
	t.Log("PASS: unsubscribe stops callbacks and leaves the other client active")

	register(first, firstEvents)

	awaitEvent(firstEvents, nodeEvent{kind: "add", name: name, after: "after-unsubscribe"})
	t.Log("PASS: resubscription replays existing state")

	if err := root.DeleteConfig(ctx); err != nil {
		t.Fatal(err)
	}
	configDeleted = true
	for _, events := range []<-chan nodeEvent{firstEvents, secondEvents} {
		awaitEvent(events, nodeEvent{kind: "delete", name: name, before: "after-unsubscribe"})
	}
	for _, reader := range []*nexusclient.Clientset{first, second} {
		waitFor("deleted node and hierarchy removal", func() bool {
			_, directErr := reader.Config().GetConfigByName(ctx, name)
			_, childErr := reader.RootRoot().GetConfig(ctx, "default")
			return apierrors.IsNotFound(directErr) && childErr != nil
		})
	}
	t.Log("PASS: SDK deletion, delete callbacks, and hierarchy removal")

	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)
		go func() { defer group.Done(); first.RootRoot().Config("default").Unsubscribe(); first.UnsubscribeAll() }()
	}
	group.Wait()
	second.UnsubscribeAll()
	if dropped.Load() != 0 {
		t.Fatal(fmt.Sprintf("event channel overflow: %d", dropped.Load()))
	}
	t.Log("PASS: repeated/concurrent unsubscribe and complete informer shutdown")
}
