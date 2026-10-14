package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/duplicated-uris-datamodel/project"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.Node
	Project project.Project `nexus:"child"`
}
