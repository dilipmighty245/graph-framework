package edge

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

var EdgeRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1alpha1/sites/{site.Site}/edges/{edge.Edge}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1alpha1/sites/{site.Site}/edges",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:EdgeRestAPISpec
// nexus-description: This is for adding a edge to inventory
// in the control plane.
type Edge struct {
	nexus.Node
	State EdgeState `nexus:"status"`
}

type EdgePowerInfo struct {
	TotalPowerAvailable uint32
	FreePowerAvailable  uint32
}

type EdgeState struct {
	PowerInfo                  EdgePowerInfo
	CurrentJob                 string
	CurrentJobPercentCompleted uint32
	CurrentJobPowerRequested   uint32
	TotalJobsProcessed         uint32
	TotalPowerServed           uint32
}
