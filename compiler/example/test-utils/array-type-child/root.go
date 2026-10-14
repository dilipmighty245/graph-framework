package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/array-type-child/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Config []config.Config `nexus:"child"` // not allowed
}
