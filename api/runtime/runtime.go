package runtime

import (
	"nexus/admin/api/authorization"
	tenantruntime "nexus/admin/api/runtime/tenant"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// Runtime tree.
type Runtime struct {
	UserCertificate authorization.UserCertificate `nexus:"children"`
	nexus.SingletonNode

	// Tenant runtime spec.
	Tenant tenantruntime.Tenant `nexus:"children"  json:"tenant,omitempty"`
}
