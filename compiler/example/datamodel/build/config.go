package build

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// BuildTestStruct struct is in build directory and should be ignored by parser
type BuildTestStruct struct {
	nexus.Node
	SomeInt int
}
