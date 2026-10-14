module github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/datamodel

go 1.16

require (
	github.com/dilipmighty245/graph-framework-for-microservices/nexus v0.0.0-20230226005144-24336fc988d5
	k8s.io/api v0.26.3
	k8s.io/apimachinery v0.26.3
)

replace github.com/dilipmighty245/graph-framework-for-microservices/nexus => ../../../nexus
