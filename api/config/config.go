package config

import (
	"nexus/admin/api/apigateway"
	"nexus/admin/api/authorization"
	tenantconfig "nexus/admin/api/config/tenant"
	"nexus/admin/api/config/user"
	"nexus/admin/api/connect"
	"nexus/admin/api/route"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// Config holds the Nexus configuration.
// Configuration in Nexus is intent-driven.
type Config struct {
	Authorization authorization.Authorization `nexus:"child"`
	nexus.SingletonNode

	// Gateway configuration.
	ApiGateway apigateway.ApiGateway `nexus:"child"`

	// API extensions configuration.
	Routes route.Route `nexus:"children"`

	// Exposed server configuration.
	Servers route.Server `nexus:"children"`

	// Nexus Connect configuration.
	Connect      connect.Connect     `nexus:"child"`
	Tenant       tenantconfig.Tenant `nexus:"children" json:"tenant,omitempty"`
	TenantPolicy tenantconfig.Policy `nexus:"children" json:"tenant_policy,omitempty"`

	User user.User `nexus:"children" json:"user,omitempty"`
}
