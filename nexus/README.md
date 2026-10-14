# Nexus datamodel types

This module supplies the Go declarations used to describe a Nexus datamodel. The compiler reads these declarations to generate resources and APIs.

## Node and API declarations

Import the DSL types from:

```go
import "github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
```

Embed `nexus.Node` for a graph node or `nexus.SingletonNode` for a singleton. Relationship tags define children and links. Additional types describe REST endpoints, HTTP responses, GraphQL queries, and secrets.

These markers describe the model. Runtime object operations are provided by the client generated for that model.

## Protocol definitions

`proto/query-manager` contains a query-manager protocol definition. The Makefile's `generate` target invokes protobuf generators and mock generation; install compatible versions of `protoc` and its Go plugins before using it.

Start with the [DSL reference](../compiler/DSL.md) and the [compiler examples](../compiler/example/datamodel). The module currently declares Go 1.22; toolchain changes must be coordinated with generators and their consumers.
