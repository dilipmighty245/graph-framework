package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

var LeaderRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/leader/{root.Leader}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
	},
}

// nexus-rest-api-gen:LeaderRestAPI
type Leader struct {
	nexus.Node

	Name        string
	Designation string
}
