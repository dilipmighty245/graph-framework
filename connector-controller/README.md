# Nexus connector controller

This controller watches NexusEndpoint resources and manages the connector deployments that serve them. It creates or updates deployment, service, and configuration resources for endpoint instances.

## Runtime dependencies

The controller needs Kubernetes access to the target namespace, generated admin API packages, and a connector image. `NAMESPACE` selects its runtime namespace and `NEXUS_CONNECTOR_VERSION` provides the connector image setting used by the controller.

## Development

Source is organized under `controllers`, `helpers`, and `cmd/controller`. Generate the [admin datamodel](../api/README.md) before building. The controller still references generated resource groups that must match the chosen admin model.

After resolving generated dependencies, build from this directory:

```sh
go build ./cmd/controller
```

Review `config`, the `controller` Helm chart, and `Makefile` for packaging assumptions. Test endpoint creation, updates, and deletion against a runtime when changing resource lifecycle behavior.

See [connector](../connector/README.md) for the replication service it manages.
