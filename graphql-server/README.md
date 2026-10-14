# Nexus GraphQL server host

This component watches installed Datamodel resources and loads their generated GraphQL server plugins. It connects generated GraphQL artifacts to the runtime's model lifecycle.

## Plugin lifecycle

When an enabled datamodel supplies a GraphQL artifact URL, the controller downloads the plugin to `PLUGIN_PATH`, loads it with Go's plugin package, and calls its `StartHttpServer` entry point.

The plugin and host must be built for the same platform with the same Go toolchain version, shared dependency sources, build tags, and relevant build settings. Rebuild and test them together when changing the toolchain. A Go plugin cannot be unloaded; test replacement and restart behavior explicitly. See the [Go plugin requirements](https://pkg.go.dev/plugin) for these constraints.

## Development

The entry point is `main.go`, lifecycle handling is in `controllers/datamodel_controller.go`, and Kubernetes client setup is in `pkg/client`.

From this directory:

```sh
go build .
```

Running the service requires Kubernetes configuration and an installed Datamodel resource. End-to-end validation should generate a model, build its plugin, load it through this host, and execute a GraphQL query.

Review `values.yaml`, `templates`, and `Makefile` for image, volume, and runtime configuration. See the [compiler guide](../compiler/README.md) for generated GraphQL artifacts.
