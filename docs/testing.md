# Framework testing

Run these commands from the repository root. The project contains separate Go
modules, so root-level `go test ./...` does not cover the framework.

```sh
make test.framework            # Complete source suites across framework modules
make test.framework.compiler   # Compiler packages and commands
make test.framework.gateway    # Gateway and controller suites; generated admin API required
make test.framework.e2e        # Fresh generation, generated SDK tests, and Kind lifecycle
make test.framework.builders   # Build public Go builders; generate/test inside containers
make test.framework.all        # Source suites followed by E2E
make test.framework.race       # Source suites with the race detector
make test.framework.stability  # Repeat all checks three times with race detection
```

Set `FRAMEWORK_TEST_RUNS=5` to change the repetition count. The stability target
enables source race detection by default; `FRAMEWORK_TEST_RACE=false` disables it.
Each pass uses
`-count=1` and a new test process; this also avoids Ginkgo v1's process-global
suite state when repeating tests. Source checks use `-mod=readonly`, enable Go's
package-test shuffling, and report every failing check. Failures are never retried
into success or accepted through a baseline allowlist.

## Prerequisites

Use Go 1.27, Python 3, and a platform C toolchain for race tests. Generator tooling
requires Go 1.26 or newer because the source-analysis dependency must understand
the compiler's export format. Tests need permission to start local HTTP servers.
On macOS the scripts use the external linker; `FRAMEWORK_TEST_LDFLAGS` overrides
that setting.

E2E also needs a running Docker daemon, Kind, kubectl, Git, Make, jq, Mike Farah's
yq v4, and goimports on `PATH`. It builds the framework's GraphQL and OpenAPI tools
from the working sources and clones the Kubernetes code generator at the version
pinned in `compiler/manifest.yaml`. An initial run needs public dependency and
container-image downloads.

A compatible local compiler image can supply the build tools instead:

```sh
E2E_COMPILER_IMAGE=your-local-compiler:your-tag make test.framework.e2e
```

The image must include Go, Git, Make, jq, yq v4, and goimports. The runner builds
gqlgen and nexus-openapi-gen from the current source snapshot. The host build is
the default; it avoids depending on a prepublished framework image.

`make test.framework.builders` builds `compiler/builder/Dockerfile` and
`authz-controller/builder/Dockerfile` using the official public
[`docker.io/library/golang:1.27.2-bookworm`](https://hub.docker.com/_/golang) base.
It checks their toolchains, generates both models inside the compiler builder,
and runs authorization race tests inside the authorization builder. The remaining
Kind resource, subscription, gateway, and GraphQL checks use those fresh outputs.
The builder images stay available locally for reuse. Override their tags with
`COMPILER_BUILDER_IMAGE` and `AUTHZ_BUILDER_IMAGE`.

## Coverage

The source target covers the SDK, common library, admin source model, validation,
install validator, CLI, calibration, connector, GraphQL host, full OpenAPI and
GraphQL generator forks, compiler packages and commands, kubebuilder packages
and CLI, and the file-upload service. It also checks the gateway module graph and
the test runner's isolation and cleanup behavior.

The gateway target runs its command, package, and internal tests, plus the
authorization and connector-controller suites. Generate the admin API before
using this target directly. The E2E target performs generation itself in a
snapshot of the current working tree, including uncommitted changes.

E2E performs these steps:

1. Generate the example CRDs, resource types, and clients from the source model.
2. Compile the SDK, test concurrent graph-cache updates, and run generated-client
   lifecycle and status tests with the race detector.
3. Generate the admin model and compile the selected services. Run gateway and
   controller source tests when gateway coverage is selected.
4. Create a uniquely named Kind cluster with its own kubeconfig.
5. Install the fresh example CRDs; create, read, patch, and delete graph resources.
   Check that Kubernetes rejects duplicates in lists declared unique. Run gateway
   controller tests against the cluster with the race detector.
6. Deploy the gateway and generated standalone GraphQL server. Check that both
   return the expected resource data and updated field value.
7. Delete the cluster, service image tags, port forwards, and temporary sources
   on success or failure.

Select services with `E2E_COMPONENTS=resources`, `resources,gateway`, or
`resources,graphql`. The default is `resources,gateway,graphql`.
`E2E_KIND_IMAGE` selects a Kind node image. Missing prerequisites and failed
service readiness fail the run.

The real generated-SDK subscription test runs three times in fresh Go test
processes by default (`E2E_SDK_RUNS` overrides the count). It uses the disposable
cluster and verifies:

- SDK creation of a root and child, hashed names, and hierarchy labels/caches.
- Add, update, and delete events, including old/new field values, on two clients.
- Per-client subscription isolation and safe copies of cached read objects.
- Unsubscribe stopping one observer while another continues receiving events.
- Resubscribe replaying current state, deletion clearing hierarchy caches, and
  repeated/concurrent unsubscribe completing without races or double closes.

The test is under `compiler/example/tests/integration`, enabled with the
`integration` build tag. Run it through the E2E targets so it receives freshly
generated code and an isolated cluster; it is not part of the ordinary fake-client
unit suite. Per-pass event results are saved as `subscriptions-N.log`.

The Kind service account has cluster-admin access within the disposable cluster.
The runner does not change the active kubeconfig or install into an existing
cluster. The generated standalone GraphQL server is covered; plugin replacement,
external OIDC providers, Envoy routing, and cross-cluster replication need their
own integration fixtures. The upstream kubebuilder scaffolding cluster suites
are separate from framework lifecycle tests; run their documented targets in
`kubebuilder/Makefile` when changing those plugins.

## Results

Each invocation writes logs to a unique directory under
`artifacts/framework-tests`, ignored by Git. Set `FRAMEWORK_TEST_RESULTS` to use a
different parent directory. Source runs write a toolchain log, per-check logs,
and `summary.tsv`. E2E records generation, CRDs, resource snapshots, service
builds, and probes; failures collect pod, event, and service logs before cleanup.

If cluster deletion fails, the runner reports its exact name and fails. Inspect
the first failing stage before rerunning. Passing repeated checks gives evidence
of stability for these paths; it cannot establish that every deployment and
external integration is free of flakes.
