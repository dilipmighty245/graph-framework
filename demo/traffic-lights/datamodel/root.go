package root

import (
	"trafficlight/config"
	"trafficlight/desiredconfig"
	"trafficlight/inventory"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Root struct {
	nexus.SingletonNode

	Config        config.Config               `nexus:"child"`
	DesiredConfig desiredconfig.DesiredConfig `nexus:"child"`
	Inventory     inventory.Inventory         `nexus:"child"`
}
