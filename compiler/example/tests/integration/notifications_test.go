//go:build integration

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/config.nexus.com/v1"
	gnsv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/gns.nexus.com/v1"
	nexusclient "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/nexus-client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type notificationTransport struct {
	next        http.RoundTripper
	initial     chan struct{}
	release     chan struct{}
	first       sync.Once
	parentReads *int32
}

// Hold an add-only listener on its first replay, delete both children while its
// second replay is queued, then drain the listener. User replay must not restore
// a relationship that the internal handler has already removed.
func TestDelayedAddReplayCannotRestoreDeletedChildren(t *testing.T) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		t.Fatal("KUBECONFIG must identify the disposable framework test cluster")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	writer, err := nexusclient.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	text := gnsv1.MyStr("replay")
	parent, err := writer.RootRoot().AddConfig(ctx, &configv1.Config{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
		Spec: configv1.ConfigSpec{
			MyStr0: &text, MyStr1: []gnsv1.MyStr{}, MyStr2: map[string]gnsv1.MyStr{},
			XYZPort: 8080, ABCHost: []gnsv1.Host{}, ClusterNamespaces: []configv1.ClusterNamespace{},
			TestValMarkers: configv1.TestValMarkers{MyStr: "ab", MyInt: 3, MySlice: []string{"a", "b"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if err := parent.Delete(cleanupCtx); err != nil {
			t.Error(err)
		}
	})
	for _, name := range []string{"replay-one", "replay-two"} {
		if _, err := parent.AddFooExample(ctx, &configv1.FooTypeABC{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       configv1.FooTypeABCSpec{FooA: configv1.AMap{}, FooB: configv1.BArray{}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	observer, err := nexusclient.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(observer.UnsubscribeAll)
	blocked, release := make(chan struct{}), make(chan struct{})
	var releaseOnce, first sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	replayed := make(chan struct{}, 2)
	if _, err := observer.RootRoot().Config("default").FooExample("").RegisterAddCallback(func(*nexusclient.ConfigFooTypeABC) {
		first.Do(func() { close(blocked); <-release })
		replayed <- struct{}{}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked:
	case <-ctx.Done():
		t.Fatal("initial replay never reached the slow listener")
	}
	cachedParent, err := observer.Config().GetConfigByName(ctx, parent.Name)
	if err != nil {
		t.Fatal(err)
	}
	waitForChildren := func(count int) {
		t.Helper()
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		for {
			children, err := cachedParent.GetAllFooExample(ctx)
			if err == nil && len(children) == count {
				return
			}
			select {
			case <-tick.C:
			case <-ctx.Done():
				t.Fatalf("hierarchy did not reach %d children: %v", count, err)
			}
		}
	}
	waitForChildren(2)
	for _, name := range []string{"replay-one", "replay-two"} {
		if err := parent.DeleteFooExample(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	waitForChildren(0)
	unblock()
	for i := 0; i < 2; i++ {
		select {
		case <-replayed:
		case <-ctx.Done():
			t.Fatal("add replay failed to drain")
		}
	}
	// A ghost relationship makes GetAllFooExample return NotFound when it tries
	// to resolve the deleted child. Check immediately after both adds finish.
	children, err := cachedParent.GetAllFooExample(ctx)
	if err != nil || len(children) != 0 {
		t.Fatalf("delayed add restored a deleted child: count=%d err=%v", len(children), err)
	}
	t.Log("PASS: delayed add-only replay cannot restore deleted hierarchy entries")
}

func (s *notificationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	query := req.URL.Query()
	if query.Get("resourceVersion") == "0" || query.Get("sendInitialEvents") == "true" {
		s.first.Do(func() { close(s.initial) })
		select {
		case <-s.release:
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
	if req.Method == http.MethodGet && len(strings.Split(strings.Trim(req.URL.Path, "/"), "/")) == 5 {
		atomic.AddInt32(s.parentReads, 1)
	}
	return s.next.RoundTrip(req)
}

func TestNotificationsWithoutParentReads(t *testing.T) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		t.Fatal("KUBECONFIG must identify the disposable framework test cluster")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	writer, err := nexusclient.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	resource := writer.DynamicClient.Resource(schema.GroupVersionResource{Group: "config.nexus.com", Version: "v1", Resource: "footypeabcs"})
	definition := &configv1.FooTypeABC{
		TypeMeta: metav1.TypeMeta{APIVersion: "config.nexus.com/v1", Kind: "FooTypeABC"},
		ObjectMeta: metav1.ObjectMeta{Name: "notification-orphan", Labels: map[string]string{
			"roots.root.nexus.com": "missing-root", "configs.config.nexus.com": "missing-config",
			"nexus/display_name": "orphan", "nexus/is_name_hashed": "true",
		}},
		Spec: configv1.FooTypeABCSpec{FooA: configv1.AMap{"value": "original"}, FooB: configv1.BArray{}},
	}
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(definition)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resource.Create(ctx, &unstructured.Unstructured{Object: object}, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if err := resource.Delete(cleanupCtx, definition.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	})
	initial, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	var parentReads int32
	observerConfig := rest.CopyConfig(config)
	observerConfig.Wrap(func(next http.RoundTripper) http.RoundTripper {
		return &notificationTransport{next: next, initial: initial, release: release, parentReads: &parentReads}
	})
	observer, err := nexusclient.NewForConfig(observerConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(observer.UnsubscribeAll)
	t.Cleanup(unblock)
	chain := observer.RootRoot().Config("missing-config").FooExample("orphan")
	chain.Subscribe()
	select {
	case <-initial:
	case <-ctx.Done():
		t.Fatal("initial list/stream did not start")
	}
	// An unsynchronized subscription must not turn a nonempty API list into
	// a successful empty result from its still-empty cache.
	items, err := observer.Config().ListFooTypeABCs(ctx, metav1.ListOptions{})
	if err != nil || len(items) != 1 || items[0].Name != definition.Name {
		t.Fatalf("cold-cache list: count=%d err=%v", len(items), err)
	}
	waitCtx, waitCancel := context.WithTimeout(ctx, 50*time.Millisecond)
	err = observer.WaitForCacheSync(waitCtx)
	waitCancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked initialization should time out: %v", err)
	}
	events := make(chan string, 4)
	add := func(obj *nexusclient.ConfigFooTypeABC) {
		obj.Spec.FooA["value"] = "callback-copy"
		events <- "add"
	}
	remove := func(*nexusclient.ConfigFooTypeABC) { events <- "delete" }
	if _, err := chain.RegisterEventHandler(add, nil, remove); err != nil {
		t.Fatal(err)
	}
	if _, err := chain.RegisterAddCallback(add); err != nil {
		t.Fatal(err)
	}
	if _, err := chain.RegisterDeleteCallback(remove); err != nil {
		t.Fatal(err)
	}
	unblock()
	if err := observer.WaitForCacheSync(ctx); err != nil {
		t.Fatal(err)
	}
	await := func(want string) {
		t.Helper()
		for i := 0; i < 2; i++ {
			select {
			case got := <-events:
				if got != want {
					t.Fatalf("got %s, want %s", got, want)
				}
			case <-ctx.Done():
				t.Fatalf("missing %s notification with an absent parent", want)
			}
		}
	}
	await("add")
	cached, err := observer.Config().GetFooTypeABCByName(ctx, definition.Name)
	if err != nil || cached.Spec.FooA["value"] != "original" {
		t.Fatalf("callback mutated cache: %v", err)
	}
	if err := resource.Delete(ctx, definition.Name, metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	await("delete")
	if got := atomic.LoadInt32(&parentReads); got != 0 {
		t.Fatalf("notifications made %d object GET requests", got)
	}
	t.Log("PASS: cold-cache list fallback, cancellable synchronization, missing-parent add/delete, safe copies, nil callbacks, and zero parent GETs")
}
