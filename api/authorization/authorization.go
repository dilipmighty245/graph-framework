// Package authorization defines the admin graph's Kubernetes RBAC intent.
package authorization

import "github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"

type Authorization struct {
	nexus.SingletonNode
	ResourceRole        ResourceRole        `nexus:"children"`
	ResourceRoleBinding ResourceRoleBinding `nexus:"children"`
	InstanceRole        InstanceRole        `nexus:"children"`
	InstanceRoleBinding InstanceRoleBinding `nexus:"children"`
	Users               User                `nexus:"children"`
	Groups              Group               `nexus:"children"`
}

// Verb is a Kubernetes RBAC operation, such as get, list, create or delete.
type Verb string

type ResourceType struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

type ResourceRule struct {
	Resource     ResourceType `json:"resource"`
	Verbs        []Verb       `json:"verbs"`
	Hierarchical bool         `json:"hierarchical,omitempty"`
}

type ResourceRole struct {
	nexus.Node
	Rules []ResourceRule `json:"rules"`
}

type ResourceInstance struct {
	Resource ResourceType `json:"resource"`
	Name     string       `json:"name"`
}

type InstanceRule struct {
	Instance ResourceInstance `json:"instance"`
	Verbs    []Verb           `json:"verbs"`
}

type InstanceRole struct {
	nexus.Node
	Rules []InstanceRule `json:"rules"`
}

type User struct{ nexus.Node }

type Group struct {
	nexus.Node
	Users User `nexus:"links"`
}

type ResourceRoleBinding struct {
	nexus.Node
	Role   ResourceRole `nexus:"link"`
	Users  User         `nexus:"links"`
	Groups Group        `nexus:"links"`
}

type InstanceRoleBinding struct {
	nexus.Node
	Role   InstanceRole `nexus:"link"`
	Users  User         `nexus:"links"`
	Groups Group        `nexus:"links"`
}

var CertificateSecretSpec = nexus.SecretSpec{}

// UserCertificate contains a generated user credential. Limit access through RBAC.
// nexus-secret-spec:CertificateSecretSpec
type UserCertificate struct {
	nexus.Node
	Key  string `json:"key"`
	Cert string `json:"cert"`
}
