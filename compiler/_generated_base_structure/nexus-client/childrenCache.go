//nolint:stylecheck,revive // Inherited from opensource.s
package nexus_client

import (
	"sort"
	"sync"
)

// Package level cache to track parent to child relationships.
// key: node crd type string of the parent.
// value: *nodeCache
type childCache struct{ parents sync.Map }

var defaultChildCache = &childCache{}

type nodeCache struct {
	// Key: Node hashed name string. Answers the question if a node exists or it has children.
	// Value: *childNodeTypeCache
	node sync.Map
}

type childNodeTypeCache struct {
	// Key: node crd type string of the child.
	// Value: *childrenCache
	childType sync.Map
}

type childrenCache struct {
	// Key: Node hashed name string. Answers the question if a child with the name exists.
	// Value: struct{}. Indicates presence.
	children sync.Map
}

func (c *childCache) AddChild(parentNodeType, parentName, childNodeType, childName string) {
	parentValue, _ := c.parents.LoadOrStore(parentNodeType, &nodeCache{})
	parent := parentValue.(*nodeCache)
	typeValue, _ := parent.node.LoadOrStore(parentName, &childNodeTypeCache{})
	childType := typeValue.(*childNodeTypeCache)
	cacheValue, _ := childType.childType.LoadOrStore(childNodeType, &childrenCache{})
	cacheValue.(*childrenCache).children.Store(childName, struct{}{})
}

func (c *childCache) RemoveChild(parentNodeType, parentName, childNodeType, childName string) {
	var typeAsserted bool

	// Remove entry on the last child being removed.
	nodeVal, parentNodeTypeFound := c.parents.Load(parentNodeType)
	if !parentNodeTypeFound {
		return
	}
	parentNodeCache, typeAsserted := (nodeVal).(*nodeCache)
	if !typeAsserted {
		logger.Fatalf("invalid nodeVal found in nodeCache for parentNodeType %s", parentNodeType)
	}

	childNodeTypeVal, parentNodeFound := parentNodeCache.node.Load(parentName)
	if !parentNodeFound {
		return
	}
	childNodeTypeCache, typeAsserted := childNodeTypeVal.(*childNodeTypeCache)
	if !typeAsserted {
		logger.Fatalf("invalid childNodeTypeVal found in childNodeTypeVal for parentName %s", parentName)
	}

	childrenCacheVal, childNodeTypeFound := childNodeTypeCache.childType.Load(childNodeType)
	if !childNodeTypeFound {
		return
	}
	childrenCache, typeAsserted := childrenCacheVal.(*childrenCache)
	if !typeAsserted {
		logger.Fatalf("invalid childrenCacheVal found in childrenCache for childNodeType %s", childNodeType)
	}

	childrenCache.children.Delete(childName)
}

func (c *childCache) IsChildExists(parentNodeType, parentName, childNodeType, childName string) bool {
	var typeAsserted bool

	nodeVal, parentNodeTypeFound := c.parents.Load(parentNodeType)
	if !parentNodeTypeFound {
		return false
	}
	parentNodeCache, typeAsserted := (nodeVal).(*nodeCache)
	if !typeAsserted {
		logger.Fatalf("invalid nodeVal found in nodeCache for parentNodeType %s", parentNodeType)
	}

	childNodeTypeVal, parentNodeFound := parentNodeCache.node.Load(parentName)
	if !parentNodeFound {
		return false
	}
	childNodeTypeCache, typeAsserted := childNodeTypeVal.(*childNodeTypeCache)
	if !typeAsserted {
		logger.Fatalf("invalid childNodeTypeVal found in childNodeTypeCache for parentName %s", parentName)
	}

	childrenCacheVal, childNodeTypeFound := childNodeTypeCache.childType.Load(childNodeType)
	if !childNodeTypeFound {
		return false
	}
	childrenCache, typeAsserted := childrenCacheVal.(*childrenCache)
	if !typeAsserted {
		logger.Fatalf("invalid childrenCacheVal found in childrenCache for childNodeType %s", childNodeType)
	}

	if _, ok := childrenCache.children.Load(childName); ok {
		return true
	}
	return false
}

func (c *childCache) GetChildren(parentNodeType, parentName, childNodeType string) (children []string) {
	var typeAsserted bool

	nodeVal, parentNodeTypeFound := c.parents.Load(parentNodeType)
	if !parentNodeTypeFound {
		return children
	}
	parentNodeCache, typeAsserted := (nodeVal).(*nodeCache)
	if !typeAsserted {
		logger.Fatalf("invalid nodeVal found in nodeCache for parentNodeType %s", parentNodeType)
	}

	childNodeTypeVal, parentNodeFound := parentNodeCache.node.Load(parentName)
	if !parentNodeFound {
		return children
	}
	childNodeTypeCache, typeAsserted := childNodeTypeVal.(*childNodeTypeCache)
	if !typeAsserted {
		logger.Fatalf("invalid childNodeTypeVal found in childNodeTypeCache for parentName %s", parentName)
	}

	childrenCacheVal, childNodeTypeFound := childNodeTypeCache.childType.Load(childNodeType)
	if !childNodeTypeFound {
		return children
	}
	childrenCache, typeAsserted := childrenCacheVal.(*childrenCache)
	if !typeAsserted {
		logger.Fatalf("invalid childrenCacheVal found in childrenCache for childNodeType %s", childNodeType)
	}

	childrenCache.children.Range(func(k, v interface{}) bool {
		children = append(children, k.(string))
		return true
	})
	sort.Strings(children)
	return children
}

// Package helpers operate on the legacy shared cache. Generated clients keep a
// separate cache so independent API backends cannot leak graph relationships.
func AddChild(parentType, parent, childType, child string) {
	defaultChildCache.AddChild(parentType, parent, childType, child)
}
func RemoveChild(parentType, parent, childType, child string) {
	defaultChildCache.RemoveChild(parentType, parent, childType, child)
}
func IsChildExists(parentType, parent, childType, child string) bool {
	return defaultChildCache.IsChildExists(parentType, parent, childType, child)
}
func GetChildren(parentType, parent, childType string) []string {
	return defaultChildCache.GetChildren(parentType, parent, childType)
}
