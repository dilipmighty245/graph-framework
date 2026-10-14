package desiredconfig

import (
	"trafficlight/desiredconfig/desiredconfiglight"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type DesiredConfig struct {
	nexus.SingletonNode

	Lights desiredconfiglight.LightConfig `nexus:"children"`
}
