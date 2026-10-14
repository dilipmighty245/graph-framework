package project

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/duplicated-uris-datamodel/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

var ProjectRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/root/{root.Root}/project/{project.Project}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
	},
}

// nexus-rest-api-gen:ProjectRestAPISpec
type Project struct {
	nexus.SingletonNode
	Key    string
	Config config.Config `nexus:"child"`
}
