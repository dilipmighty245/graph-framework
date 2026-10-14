package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/pointer-type-datamodel/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Config *config.Config `nexus:"child"` // not allowed
}
