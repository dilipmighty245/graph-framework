package light

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Light struct {
	nexus.Node

	GroupNumber int
	LightNumber int
}
