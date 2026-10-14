package root

import (
	"helloworld/config"
	"helloworld/inventory"
	"helloworld/runtime"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	MyInt     int
	Config    config.Config       `nexus:"child"`
	Runtime   runtime.Runtime     `nexus:"child"`
	Inventory inventory.Inventory `nexus:"child"`
}
