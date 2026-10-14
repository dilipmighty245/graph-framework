package sitedc

import (
	"powerschedulermodel/runtimedesiredconfig/sitedc/edgedc"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

var SiteDCRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1alpha1/info/sites/{sitedc.SiteDC}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1alpha1/info/sites",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:SiteDCRestAPISpec
// nexus-description: This is for adding a desired edge config
type SiteDC struct {
	nexus.Node

	EdgesDC edgedc.EdgeDC `nexus:"children"`
}
