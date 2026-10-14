package config

import (
	"trafficlight/common"
	"trafficlight/config/group"
	"trafficlight/config/perftest"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

var ConfigRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/config",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
	},
}

// nexus-rest-api-gen:ConfigRestAPISpec
type Config struct {
	nexus.SingletonNode

	DesiredLightAll  string
	LightDuration    common.LightDuration
	AutoReportStatus bool

	Group    group.Group       `nexus:"children"`
	PerfTest perftest.PerfTest `nexus:"child"`
}
