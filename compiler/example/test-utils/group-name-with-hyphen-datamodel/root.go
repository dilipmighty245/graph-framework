package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/common-library/pkg/nexus"
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/group-name-with-hyphen-datamodel/project"
)

type Root struct {
	nexus.Node
	SomeRootData string
	Project      project.Project `nexus:"child"`
}
