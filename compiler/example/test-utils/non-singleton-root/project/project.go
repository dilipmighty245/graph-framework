package project

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/non-singleton-root/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Project struct {
	nexus.SingletonNode
	Key    string
	Config config.Config `nexus:"child"`
}
