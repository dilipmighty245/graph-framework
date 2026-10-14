package perftest

import "github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"

var PerfTestRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/config/perftest",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
	},
}

// nexus-rest-api-gen:PerfTestRestAPISpec
type PerfTest struct {
	nexus.SingletonNode

	// possible values if not empty: Broadcast, Unicast, Boot
	Mode           string
	TestInProgress bool
}
