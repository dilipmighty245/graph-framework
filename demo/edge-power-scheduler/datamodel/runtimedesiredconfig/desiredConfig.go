package runtimedesiredconfig

import (
	"powerschedulermodel/runtimedesiredconfig/sitedc"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

type DesiredSiteConfig struct {
	nexus.SingletonNode
	SitesDC sitedc.SiteDC `nexus:"children"`
}
