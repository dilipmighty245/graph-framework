package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/invalid-type-name-datamodel/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Config config.Config `nexus:"child"`
}
