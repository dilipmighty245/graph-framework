package edgedc

import (
	"powerschedulermodel/runtimedesiredconfig/sitedc/edgedc/jobmgmt"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

var EdgeDCRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1alpha1/info/sites/{sitedc.SiteDC}/edge/{edgedc.EdgeDC}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1alpha1/info/sites/{sitedc.SiteDC}/edge",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:EdgeDCRestAPISpec
// nexus-description: This is for adding a desired edge config
type EdgeDC struct {
	nexus.Node

	JobsInfo jobmgmt.JobInfo `nexus:"children"`
}
