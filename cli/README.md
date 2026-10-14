# Nexus CLI

The Nexus CLI coordinates datamodel generation, application scaffolding, and runtime operations. It delegates work to component Makefiles, container images, Kubernetes tools, and downloaded templates.

## Commands

| Command group | Purpose |
| --- | --- |
| `datamodel` | Initialize, build, package, install, and configure datamodels. |
| `app` | Initialize applications, attach datamodels, build images, and deploy. |
| `runtime` | Install and manage the runtime environment. |
| `operator` | Scaffold controllers for datamodel resources. |
| `prereq` | Inspect workflow prerequisites. |
| `config`, `login`, `apply`, `delete`, `debug`, `version`, `upgrade` | Manage configuration, access, resources, and diagnostics. |

Use command help to inspect flags and subcommands for your checkout.

## Build from source

From this directory:

```sh
go build -o ./artifacts/nexus/nexus ./cmd/plugin/nexus
./artifacts/nexus/nexus --help
```

This builds for the local operating system and architecture. The existing Makefile also has platform-specific targets; inspect their architecture settings before using them.

The source contains a separate TSM entry point under `cmd/plugin/tsm`. The Nexus entry point is `cmd/plugin/nexus`.

## Workflow

A typical application workflow defines a datamodel, runs the compiler, installs the generated CRDs, and builds a service using the generated Go client. See the [framework overview](../README.md), [compiler guide](../compiler/README.md), and [template guide](../datamodel-templates/README.md).

Runtime and template commands may download artifacts and modify a Kubernetes environment. Configure image registries, artifact versions, namespaces, and kubeconfig explicitly. Several defaults refer to historical deployment environments.

## Development

This is a separate Go module. Run compilation and vet checks here. The Makefile's `test` target is a placeholder, so it is not evidence of command coverage. Validate affected commands and the build or deployment operations they invoke.
