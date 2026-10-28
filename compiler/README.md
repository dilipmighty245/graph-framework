# Nexus compiler

The compiler turns a graph-shaped Go datamodel into Kubernetes resources and application-facing APIs. It is the bridge between the [Nexus DSL](DSL.md) and the runtime.

## Inputs and outputs

Inputs are Go structs containing Nexus node markers, relationship tags, API declarations, and a model configuration file. Outputs include:

- CRD manifests and OpenAPI schemas.
- Kubernetes resource types, deepcopy implementations, and Go clients.
- A Nexus client layer for parent–child operations, links, naming, and subscriptions.
- GraphQL schemas, resolvers, and server sources.
- Installation and supporting artifacts used by downstream applications.

The pipeline parses the model, builds its relationship graph, renders templates, and runs Kubernetes and GraphQL generators. See [example/datamodel](example/datamodel) for input. Generated example output is ignored by Git and created on demand.

Keep the example datamodel's Go source, GraphQL queries, module files, and
configuration in Git: compiler tests and generation use them as inputs. Its
`build/` directory contains generated clients, informers, CRDs, and GraphQL
output; it is ignored along with `example/output/`. The DSL parser skips `build/`
entirely, so these artifacts are not required in a checkout.

## Repository dependencies

Keep this directory beside `nexus`, `common-library`, `kube-openapi`, `gqlgen`, and `install-validator`. The module and build scripts use local paths to these components. Kubernetes code-generator is recorded as a repository submodule.

From this directory, initialize it with:

```sh
make init_submodules
```

Generation also uses tools including `protoc`, Go protobuf generators, `yq`, `jq`, and Kubernetes generators. Inspect `Makefile`, `builder/Dockerfile`, `manifest.yaml`, and `scripts/` before provisioning a build environment. Existing container builders need toolchain and base-image alignment before they can be treated as a reproducible default.

## Kubernetes compatibility

The compiler and generated clients use Kubernetes v0.34.3 and Go 1.24 or newer.
The code-generator submodule and `manifest.yaml` pin the same release. Generation
uses `kube_codegen.sh`; the removed `generate-groups.sh` path is no longer used.
Regenerate existing datamodel clients and rebuild their services and GraphQL
plugins when adopting this release.

For streaming initialization, set `KUBE_FEATURE_WatchListClient=true` in the
application's environment before it starts. The API server must support WatchList
(the validated target is v1.34.3). Without that variable, client-go retains its
LIST/WATCH default. Failed streaming initialization falls back through client-go.
The runtime compatibility notes are in
[the runtime guide](../nexus-runtime-manifests/README.md).

## Generation

The principal Make inputs are `DATAMODEL_PATH`, `CONFIG_FILE`, `GENERATED_OUTPUT_DIRECTORY`, and `CRD_MODULE_PATH`. After preparing the toolchain, the included example can be generated with:

```sh
make generate_example
```

That target writes under `example/output/generated`. These files are local build artifacts and should not be committed. To inspect the parser entry point and its options:

```sh
go run ./cmd/nexus-sdk -help
```

The CLI wraps these operations for application workflows; see [cli](../cli/README.md).

## Validation

Run package checks from this module. Parser, preparser, generator, and generated-client tests cover different stages.

Generator unit tests compare against the seven regression fixtures in `pkg/generator/testdata`; they run without generated example output. `example/tests` and other nested modules need their own checks; `go test ./...` in this module does not traverse module boundaries. From the repository root, `make test.framework.e2e` regenerates the example in an isolated snapshot, compiles the result, and exercises its APIs in a fresh Kind cluster. `make clean` removes locally generated example output.
