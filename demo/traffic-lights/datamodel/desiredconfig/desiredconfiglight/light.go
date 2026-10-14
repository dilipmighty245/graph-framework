package desiredconfiglight

import (
	"trafficlight/common"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

var DesiredConfigLightRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/desiredconfig/light/{desiredconfiglight.LightConfig}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/desiredconfig/lights",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:DesiredConfigLightRestAPISpec
type LightConfig struct {
	nexus.Node

	GroupID                       int
	LightID                       int
	DesiredLightAll               string `json:"desiredLightAll,omitempty" mapstructure:"desiredLightAll,omitempty"`
	LightDuration                 common.LightDuration
	AutoReportStatus              bool
	OnboardTime                   string
	LastManualStatusRequestedTime string `json:"lastManualStatusRequestedTime,omitempty" mapstructure:"lastManualStatusRequestedTime,omitempty"`
	ConfigInfo                    LightConfigStatus
	Status                        LightStatus `nexus:"children"`
}

type LightStatus struct {
	nexus.Node
	Config           LightConfigStatus
	AutoStatus       LightInfo
	ManualStatus     LightInfo
	BroadcastLatency Latency
	UnicastLatency   Latency
	Boottime         Boottime
}

type LightInfo struct {
	CurrentLight         string
	LastStatusReportTime string
	NumRedLights         uint64
	NumYellowLights      uint64
	NumGreenLights       uint64
}

type LightConfigStatus struct {
	Version   string
	Timestamp string
}

type Boottime struct {
	ObjectCount int64
	Latency     Latency
}

type Latency struct {
	Max int64
	Min int64
}
