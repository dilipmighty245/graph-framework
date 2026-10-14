package root

import (
	global "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/global-package/project"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Project global.Project `nexus:"child"`
	IsRoot  IsRoot         // <--- to verify alias type
}

type IsRoot bool
