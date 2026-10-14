package inventory

import (
	"powerschedulermodel/inventory/site"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

type Inventory struct {
	nexus.SingletonNode
	Site site.Site `nexus:"children"`
}
