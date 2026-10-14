package api

import (
	"nexus/admin/api/config"
	"nexus/admin/api/runtime"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// Nexus is the root node for Nexus infra/runtime datamodel.
//
// This hosts the graph that will consist of user configuration,
// runtime state, inventory and other state essential to the
// functioning of Nexus SDK and runtime.
type Nexus struct {
	nexus.SingletonNode

	// Configuration.
	Config  config.Config   `nexus:"child"`
	Runtime runtime.Runtime `nexus:"child" json:"runtime,omitempty"`
}
