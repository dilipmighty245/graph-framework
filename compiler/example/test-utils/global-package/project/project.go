package global

import (
	global "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/global-package/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Project struct {
	nexus.SingletonNode
	Key    string
	Config global.Config `nexus:"child"`
}
