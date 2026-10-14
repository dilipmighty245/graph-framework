package nexus_client

import (
	"fmt"
	"sync"
	"testing"
)

func TestChildCacheConcurrentInsertion(t *testing.T) {
	cache := &childCache{}
	var workers sync.WaitGroup
	for i := 0; i < 100; i++ {
		workers.Add(1)
		go func(i int) { defer workers.Done(); cache.AddChild("parent", "one", "child", fmt.Sprint(i)) }(i)
	}
	workers.Wait()
	if got := len(cache.GetChildren("parent", "one", "child")); got != 100 {
		t.Fatalf("lost children: got %d, want 100", got)
	}
	if other := (&childCache{}).GetChildren("parent", "one", "child"); len(other) != 0 {
		t.Fatal("relationships leaked between caches")
	}
}
