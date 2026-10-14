package servicegroup

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

// nexus-secret-spec:ApiKeySecretSpec
type SvcGroup struct {
	nexus.Node
	DisplayName string
	Description string
	Color       string
}

type SvcGroupLinkInfo struct {
	nexus.Node
	ClusterName string
	DomainName  string
	ServiceName string
	ServiceType string
}
