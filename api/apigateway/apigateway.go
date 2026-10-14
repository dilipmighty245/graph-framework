package apigateway

import (
	"nexus/admin/api/admin"
	authentication "nexus/admin/api/authn"
	domain "nexus/admin/api/domain"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// ApiGateway holds all configuration relevant to a gateway in Nexus runtime.
type ApiGateway struct {
	nexus.Node

	// ProxyRules define a match condition and a corresponding upstream
	ProxyRules admin.ProxyRule `nexus:"children"`

	// Authentication config associated with this Gateway.
	Authn authentication.OIDC `nexus:"child"`

	//Domain objects
	Cors domain.CORSConfig `nexus:"children"`
}
