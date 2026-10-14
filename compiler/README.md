# Nexus compiler

The compiler turns a graph-shaped Go datamodel into Kubernetes resources and application-facing APIs. It is the bridge between the [Nexus DSL](DSL.md) and the runtime.

## Inputs and outputs

Inputs are Go structs containing Nexus node markers, relationship tags, API declarations, and a model configuration file. Outputs include:

- CRD manifests and OpenAPI schemas.
- Kubernetes resource types, deepcopy implementations, and Go clients.
- A Nexus client layer for parent–child operations, links, naming, and subscriptions.
- GraphQL schemas, resolvers, and server sources.
- Installation and supporting artifacts used by downstream applications.

The pipeline parses the model, builds its relationship graph, renders templates, and runs Kubernetes and GraphQL generators. See [example/datamodel](example/datamodel) for input and [example/output](example/output) for checked-in output and fixtures.

## Repository dependencies

Keep this directory beside `nexus`, `common-library`, `kube-openapi`, `gqlgen`, and `install-validator`. The module and build scripts use local paths to these components. Kubernetes code-generator is recorded as a repository submodule.

From this directory, initialize it with:

```sh
make init_submodules
```

Generation also uses tools including `protoc`, Go protobuf generators, `yq`, `jq`, and Kubernetes generators. Inspect `Makefile`, `builder/Dockerfile`, `manifest.yaml`, and `scripts/` before provisioning a build environment. Existing container builders need toolchain and base-image alignment before they can be treated as a reproducible default.

## Generation

The principal Make inputs are `DATAMODEL_PATH`, `CONFIG_FILE`, `GENERATED_OUTPUT_DIRECTORY`, and `CRD_MODULE_PATH`. After preparing the toolchain, the included example can be generated with:

```sh
make generate_example
```

That target writes under `example/output/generated`. Run it in a clean checkout and review output changes. To inspect the parser entry point and its options:

```sh
go run ./cmd/nexus-sdk -help
```

The CLI wraps these operations for application workflows; see [cli](../cli/README.md).

## Validation

Run package checks from this module. Parser, preparser, generator, and generated-client tests cover different stages. Several checked-in fixtures currently disagree with the example model, so establish the baseline before attributing failures to a change.

Files under `example/output/_rendered_templates` are expected output read by tests. `example/tests` and other nested modules need their own checks; `go test ./...` in this module does not traverse module boundaries. End-to-end validation must regenerate a model, compile the result, and exercise its APIs against a runtime.
