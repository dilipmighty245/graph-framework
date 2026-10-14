package site

import (
	"powerschedulermodel/inventory/site/edge"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

var SiteRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1alpha1/sites/{site.Site}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1alpha1/sites",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:SiteRestAPISpec
// nexus-description: This is for adding a edge to inventory
// in the control plane.

type Site struct {
	nexus.Node
	Edges edge.Edge `nexus:"children"`
}
