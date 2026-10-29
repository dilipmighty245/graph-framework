//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/config.nexus.com/v1"
	gnsv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/gns.nexus.com/v1"
	rootv1 "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/apis/root.nexus.com/v1"
	nexusclient "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/nexus-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientfeatures "k8s.io/client-go/features"
	clientfeaturestesting "k8s.io/client-go/features/testing"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Count actual successful API requests so an enabled feature gate alone cannot
// make a streaming experiment pass after silently falling back to LIST/WATCH.
type subscriptionRequests struct {
	lists, watches, initialStreams, rejectedStreams atomic.Int64
}

type subscriptionTransport struct {
	next          http.RoundTripper
	requests      *subscriptionRequests
	rejectInitial bool
}

func (s subscriptionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.rejectInitial && strings.Contains(req.URL.Path, ".nexus.com/") &&
		req.URL.Query().Get("watch") == "true" && req.URL.Query().Get("sendInitialEvents") == "true" {
		s.requests.rejectedStreams.Add(1)
		// Simulate a server rejecting WatchList. All fallback traffic goes to
		// the real cluster; no synthetic snapshot or watch events are supplied.
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"BadRequest","message":"sendInitialEvents is unsupported","code":400}`)),
			Request:    req,
		}, nil
	}
	response, err := s.next.RoundTrip(req)
	if err == nil && response.StatusCode == http.StatusOK && strings.Contains(req.URL.Path, ".nexus.com/") {
		query := req.URL.Query()
		switch {
		case query.Get("watch") == "true" && query.Get("sendInitialEvents") == "true":
			s.requests.initialStreams.Add(1)
		case query.Get("watch") == "true":
			s.requests.watches.Add(1)
		case req.Method == http.MethodGet && len(strings.Split(strings.Trim(req.URL.Path, "/"), "/")) == 4:
			s.requests.lists.Add(1)
		}
	}
	return response, err
}

func restartParameter(t *testing.T, name string, fallback int) int {
	t.Helper()
	if value := os.Getenv(name); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			t.Fatalf("%s must be a positive integer", name)
		}
		return parsed
	}
	return fallback
}

// TestSubscriptionRestartStorm starts every subscriber against a populated
// graph, stops them all, changes the graph while offline, then starts fresh
// clients simultaneously. Run each transport mode in a fresh Go process.
func TestSubscriptionRestartStorm(t *testing.T) {
	testSubscriptionRestartStorm(t, false)
}

func TestSubscriptionWatchListFallback(t *testing.T) {
	clientfeaturestesting.SetFeatureDuringTest(t, clientfeatures.WatchListClient, true)
	t.Setenv("SUBSCRIPTION_EXPECT_WATCH_LIST", "false")
	testSubscriptionRestartStorm(t, true)
}

func testSubscriptionRestartStorm(t *testing.T, rejectInitial bool) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		t.Fatal("KUBECONFIG must identify the disposable framework test cluster")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	clientCount := restartParameter(t, "SUBSCRIPTION_RESTART_CLIENTS", 8)
	objectCount := restartParameter(t, "SUBSCRIPTION_RESTART_OBJECTS", 32)
	payloadBytes := restartParameter(t, "SUBSCRIPTION_RESTART_PAYLOAD_BYTES", 4096)
	config.QPS, config.Burst = 100, 200
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	writer, err := nexusclient.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	root, err := writer.AddRootRoot(ctx, &rootv1.Root{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if err := writer.DeleteRootRoot(cleanupCtx); err != nil {
			t.Errorf("cleaning restart graph: %v", err)
		}
	})
	text := gnsv1.MyStr("restart-test")
	parent, err := root.AddConfig(ctx, &configv1.Config{
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
	expected := map[string]bool{}
	var first, last *nexusclient.ConfigFooTypeABC
	add := func(name string) *nexusclient.ConfigFooTypeABC {
		t.Helper()
		node, err := parent.AddFooExample(ctx, &configv1.FooTypeABC{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: configv1.FooTypeABCSpec{
				FooA: configv1.AMap{"payload": strings.Repeat("x", payloadBytes), "generation": "seed"},
				FooB: configv1.BArray{},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		expected[node.Name] = true
		return node
	}
	for i := 0; i < objectCount; i++ {
		last = add(fmt.Sprintf("restart-%04d", i))
		if i == 0 {
			first = last
		}
	}
	for round := 0; round < 2; round++ {
		t.Run(fmt.Sprintf("round-%d", round+1), func(t *testing.T) {
			requests := &subscriptionRequests{}
			clients := make([]*nexusclient.Clientset, clientCount)
			for i := range clients {
				cfg := rest.CopyConfig(config)
				cfg.Wrap(func(next http.RoundTripper) http.RoundTripper {
					return subscriptionTransport{next: next, requests: requests, rejectInitial: rejectInitial}
				})
				clients[i], err = nexusclient.NewForConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(clients[i].UnsubscribeAll)
			}
			gate := make(chan struct{})
			errors := make(chan error, clientCount)
			var started sync.WaitGroup
			started.Add(clientCount)
			for i, client := range clients {
				go func(index int, client *nexusclient.Clientset) {
					defer started.Done()
					<-gate
					client.SubscribeAll()
					if err := client.WaitForCacheSync(ctx); err != nil {
						errors <- fmt.Errorf("client %d cache synchronization: %w", index, err)
						return
					}
					var seenMu sync.Mutex
					seen := map[string]bool{}
					_, err := client.RootRoot().Config("default").FooExample("").RegisterAddCallback(func(obj *nexusclient.ConfigFooTypeABC) {
						seenMu.Lock()
						seen[obj.Name] = true
						seenMu.Unlock()
					})
					if err != nil {
						errors <- fmt.Errorf("client %d registration: %w", index, err)
						return
					}
					tick := time.NewTicker(10 * time.Millisecond)
					defer tick.Stop()
					for {
						seenMu.Lock()
						complete := len(seen) == len(expected)
						for name := range expected {
							complete = complete && seen[name]
						}
						seenMu.Unlock()
						if complete {
							// Check the internal hierarchy handler too: informer sync
							// alone does not prove that the graph cache is populated.
							cachedParent, parentErr := client.RootRoot().GetConfig(ctx, "default")
							if parentErr == nil {
								children, childrenErr := cachedParent.GetAllFooExample(ctx)
								if childrenErr == nil && len(children) == len(expected) {
									return
								}
							}
						}
						select {
						case <-ctx.Done():
							errors <- fmt.Errorf("client %d initial replay/hierarchy: %w", index, ctx.Err())
							return
						case <-tick.C:
						}
					}
				}(i, client)
			}
			start := time.Now()
			close(gate)
			started.Wait()
			close(errors)
			for err := range errors {
				t.Error(err)
			}
			if t.Failed() {
				return
			}
			startup := time.Since(start)
			// Every cache must contain exactly the current snapshot, including
			// the replacement of a child while all previous subscribers were down.
			for i, client := range clients {
				items, err := client.Config().ListFooTypeABCs(ctx, metav1.ListOptions{})
				if err != nil || len(items) != len(expected) {
					t.Fatalf("client %d cached snapshot: count=%d err=%v", i, len(items), err)
				}
				for _, item := range items {
					if !expected[item.Name] {
						t.Fatalf("client %d retained unexpected child %s", i, item.Name)
					}
				}
			}
			last.Spec.FooA["generation"] = fmt.Sprintf("live-%d", round)
			if err := last.Update(ctx); err != nil {
				t.Fatal(err)
			}
			for i, client := range clients {
				for {
					item, err := client.Config().GetFooTypeABCByName(ctx, last.Name)
					if err == nil && item.Spec.FooA["generation"] == last.Spec.FooA["generation"] {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatalf("client %d missed live update: %v", i, ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
			}
			if want := os.Getenv("SUBSCRIPTION_EXPECT_WATCH_LIST"); want != "" {
				streaming, err := strconv.ParseBool(want)
				if err != nil {
					t.Fatal(err)
				}
				if streaming && (requests.initialStreams.Load() < int64(clientCount) || requests.lists.Load() != 0) {
					t.Fatalf("expected streaming without LIST fallback: streams=%d lists=%d", requests.initialStreams.Load(), requests.lists.Load())
				}
				if !streaming && (requests.initialStreams.Load() != 0 || requests.lists.Load() < int64(clientCount)) {
					t.Fatalf("expected LIST/WATCH: streams=%d lists=%d", requests.initialStreams.Load(), requests.lists.Load())
				}
			}
			if rejectInitial && (requests.rejectedStreams.Load() < int64(clientCount) || requests.watches.Load() < int64(clientCount)) {
				t.Fatalf("expected rejected WatchList followed by real LIST/WATCH: rejected=%d lists=%d watches=%d", requests.rejectedStreams.Load(), requests.lists.Load(), requests.watches.Load())
			}
			t.Logf("clients=%d objects=%d payload_bytes=%d startup=%s lists=%d watches=%d initial_streams=%d rejected_streams=%d; replay, hierarchy, snapshot and live updates verified", clientCount, len(expected), payloadBytes, startup, requests.lists.Load(), requests.watches.Load(), requests.initialStreams.Load(), requests.rejectedStreams.Load())
		})
		if t.Failed() {
			return
		}
		if round == 0 {
			if err := parent.DeleteFooExample(ctx, first.DisplayName()); err != nil {
				t.Fatal(err)
			}
			delete(expected, first.Name)
			last = add("restart-offline-replacement")
		}
	}
}
