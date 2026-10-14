module github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/group-name-with-hyphen-datamodel

go 1.18

require (
	github.com/dilipmighty245/graph-framework-for-microservices/common-library v0.0.0-20221028160844-d70f863bc31b
	github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/duplicated-uris-datamodel v0.0.0-20221028160844-d70f863bc31b
)

require github.com/elliotchance/orderedmap v1.4.0 // indirect

replace github.com/dilipmighty245/graph-framework-for-microservices/common-library => ../../../../common-library

replace github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/test-utils/duplicated-uris-datamodel => ../duplicated-uris-datamodel
