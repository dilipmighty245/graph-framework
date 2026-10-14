# Nexus resource connector

The connector replicates configured resources between endpoints. It uses NexusEndpoint and ReplicationConfig resources to select the destination, source objects or resource types, filters, and hierarchy behavior.

## Replication flow

1. Read endpoint and replication configuration.
2. Discover resource types and watch source objects.
3. Select objects using configured names, namespaces, and labels.
4. Create, update, or delete destination resources and track replication state.

Configuration can map resources between hierarchical and non-hierarchical models and between different resource types. Status replication is an additional configurable path.

## Source layout

`pkg/handlers` implements replication behavior. `pkg/app` sets up informers and controllers. `pkg/handler` dispatches events using a rate-limited queue, preserving order per object while allowing independent objects to progress concurrently. `controllers` tracks resource discovery and controller state.

## Development

From this directory:

```sh
go test -race ./pkg/handler
go build ./cmd/connector
```

The existing dependency graph includes two gnostic protobuf registrations. Broader tests use the compatibility flag already present in the Makefile:

```sh
go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" ./controllers ./pkg/handlers
```

This setting permits tests to run; it does not resolve the dependency conflict. Tests using local HTTP servers need permission to bind to localhost.

## Deployment

Inspect `pkg/config`, `values.yaml`, and the `nexus-connector` chart for endpoint, credential, and image settings. The [connector controller](../connector-controller/README.md) manages connector instances; this service performs the replication itself.
