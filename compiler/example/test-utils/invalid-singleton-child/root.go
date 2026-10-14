package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Id     string
	Config Cfg `nexus:"children"`
}

type Cfg struct {
	nexus.SingletonNode
}
