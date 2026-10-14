package config

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type Config struct {
	nexus.Node
	Id          string
	FooResource Resource
}

type Resource struct {
	Name string
}
