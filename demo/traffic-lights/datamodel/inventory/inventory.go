package inventory

import (
	"trafficlight/inventory/light"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Inventory struct {
	nexus.SingletonNode

	Lights light.Light `nexus:"children"`
}
