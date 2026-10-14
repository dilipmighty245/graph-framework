package ignored

import "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/datamodel/config/gns"

type Config struct {
	GNS gns.Gns `nexus:"child"`
}
