package group

import "github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"

type Group struct {
	nexus.Node
	Status GroupStatus `nexus:"status"`
}

type GroupStatus struct {
	TotalLights uint32
}
