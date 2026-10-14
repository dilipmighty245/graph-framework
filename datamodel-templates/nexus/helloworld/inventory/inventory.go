package inventory

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Inventory struct {
	nexus.Node
	InventoryId int
}
