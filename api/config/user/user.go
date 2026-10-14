package user

import (
	tenantconfig "nexus/admin/api/config/tenant"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type User struct {
	nexus.Node

	Username  string `json:"username" yaml:"username"`
	Mail      string `json:"email,omitempty" yaml:"mail,omitempty"`
	FirstName string `json:"firstName,omitempty" yaml:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty" yaml:"lastName,omitempty"`
	Password  string `json:"password" yaml:"password"`
	TenantId  string `json:"tenantId" yaml:"tenantId"`
	Realm     string `json:"realm" yaml:"realm,omitempty"`

	Tenant tenantconfig.Tenant `nexus:"link"`
}
