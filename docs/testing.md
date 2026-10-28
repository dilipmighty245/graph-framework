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

The example model uses the `nexus.com` base domain, including `root.nexus.com`
and `config.nexus.com`. Changing an API group creates new Kubernetes resource
types; it does not migrate resources stored under a previous group. Use a fresh
disposable cluster when testing generated example models.

`compiler/example/output/` is generated on demand and ignored by Git. The E2E
runner creates it inside its temporary source snapshot; source-only compiler
tests use regression fixtures in `compiler/pkg/generator/testdata`.

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

## Subscription restart and streaming experiment

The integration suite also starts eight independent SDK clients simultaneously
against 32 existing child objects with 4 KiB payloads. It verifies every expected
initial add callback, hierarchy-cache population, cached snapshot membership,
and live updates. It then stops every subscriber, deletes and replaces a child
while offline, and starts fresh clients together to check recovery.

Compare the existing dependencies with an upgraded client using both LIST/WATCH
and WatchList against the same disposable API server:

```sh
E2E_COMPONENTS=resources E2E_SDK_RUNS=1 \
E2E_KIND_IMAGE=kindest/node:v1.34.3 \
E2E_SUBSCRIPTION_COMPARE_STREAMING=true \
bash scripts/test-framework-e2e.sh
```

This opt-in mode upgrades Kubernetes dependencies only in the temporary test
module to v0.34.3 (`E2E_SUBSCRIPTION_CLIENT_GO_VERSION` overrides this version).
The production dependency pins and code generator are unchanged. It exercises
the existing generated clients with the upgraded informer library; it does not
establish compatibility of a repository-wide upgrade. The selected API server
must support WatchList and have it enabled.

Each mode runs in a separate Go process to isolate client-go feature gates.
Successful LIST, ordinary WATCH, and initial-event WATCH requests are counted.
Streaming must synchronize successfully with no LIST fallback. Logs are
`subscriptions-N.log`, `subscriptions-control-N.log`,
`subscriptions-streaming-N.log`, and `subscriptions-upgrade.log`.

Set `SUBSCRIPTION_RESTART_CLIENTS`, `SUBSCRIPTION_RESTART_OBJECTS`, and
`SUBSCRIPTION_RESTART_PAYLOAD_BYTES` to positive integers to vary the workload.
Startup timings are observations from a correctness test with the race detector,
not a capacity benchmark. The test now waits for every subscribed informer store
to synchronize before checking callbacks and hierarchy contents. It does not
exercise API-server restart/compaction or slow application callbacks, and does
not measure API-server peak memory.

Generated clients expose an optional, cancellable initialization wait:

```go
client.SubscribeAll()
if err := client.WaitForCacheSync(ctx); err != nil {
    return err
}
```

This waits for informer stores, not for application callbacks to finish. Existing
subscription and callback signatures are preserved. List reads use the API while
an informer is initially unsynchronized, then use its cache as before. Add/delete
notifications maintain relationships from object labels without fetching or
polling parents. Both ordinary deletions and informer deletion tombstones are
accepted. No notification queue or event-coalescing policy is added.

The notification regression test holds the initial LIST/stream open and verifies
that an immediate list still returns the API's objects, synchronization can time
out, and callbacks receive safe object copies. It then verifies both combined and
individual add/delete callbacks with absent parents and zero object GET requests.
Generated-client unit tests cover concurrent informer creation and tombstones.
The delayed-replay regression blocks an add-only listener, deletes its queued
children, then drains replay and checks that no hierarchy entries reappear.
Hierarchy state has one internal writer; application callbacks only receive
copies and never reapply hierarchy updates. See the [compiler release
notes](../compiler/CHANGELOG.md) for notification timing changes.

`compiler/_generated_base_structure/` is a generation input, not a standalone Go
module. Its tests need the emitted client definitions and must run after
generation. `make test.framework.e2e` generates an isolated SDK and runs its unit
tests before creating Kind. After generating a local example SDK, they can also
be run from `compiler/example/output/generated` with `go test -race ./nexus-client`
(include `-ldflags=-linkmode=external` on macOS).

Run compiler source tests and the E2E command locally before submitting changes.
Local template-directory test commands do not replace the generated-SDK check.

Validated after the notification refactor on 2026-10-28 with Go 1.27.0,
Kind Kubernetes v1.34.3, the race detector, and the default workload above.
Each mode ran in three fresh processes with two startup rounds per process
(18 restart rounds total):

| Client / transport | LISTs per round | Ordinary WATCHes per round | Initial-event streams per round | Startup range across six rounds |
| --- | ---: | ---: | ---: | --- |
| v0.26.3 / LIST-WATCH | 120 | 120 | 0 | 117.84–124.63 ms |
| v0.34.3 / LIST-WATCH | 120 | 120 | 0 | 115.78–125.63 ms |
| v0.34.3 / WatchList | 0 | 0 | 120 | 116.65–125.86 ms |

All nine processes passed the delayed-replay hierarchy regression, notification
regression, restart, and existing subscription lifecycle tests. Notification callbacks made zero object GET
requests in every mode. Fresh generation of both models, generated-client
unit/race tests, gateway/controller checks, resource CRUD/validation, gateway and
GraphQL probes, and all 17 framework source checks passed. The disposable cluster
was deleted. Raw logs are in `artifacts/framework-tests/e2e-eJmxzGU2/` and
`artifacts/framework-tests/run-P5NovJ6X/` (ignored by Git).

The generated SDK can use streaming with the upgraded informer library in this
experiment. These startup times include the explicit cache-synchronization wait
and are not directly comparable with the earlier experiment. They do not
establish a latency or memory improvement. Production dependency pins remain
unchanged; adopting streaming still requires a coordinated dependency/runtime
upgrade. Slow application callbacks remain the caller's responsibility.

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

## Cleanup

```sh
make clean.preview                    # Inspect project cleanup without changing resources
make clean                            # Remove project runtime and generated artifacts
make clean CLUSTER_NAME=my-cluster     # Include a specifically named Kind cluster
make clean.all                        # Include ALL local Docker-backed Kind clusters
python3 scripts/clean-framework.py --all-kind-clusters --dry-run
```

Cleanup needs Python 3, Git, Docker, and Kind when matching clusters exist. Stop
active builds and test runs first. Project clusters are identified by directories
under `nexus-runtime-manifests/kind/` or the test runner's `nexus-e2e-` prefix.
Cluster deletion also removes its Kubernetes workloads and node storage.

Cleanup removes associated proxy/gateway containers, this checkout's legacy
Compose runtime containers and their anonymous volumes, the legacy k0s kubeconfig
volume, unused `nexus`/`kind` networks, and local image tags under `nexus/`,
`nexus-e2e/`, `authz`, and `authz-builder`. Image removal is not forced; images
still used by other containers cause a reported failure. Registry-prefixed or
custom image tags are retained.

Local cleanup removes build/test artifacts, generated admin APIs, executables,
coverage output, Python bytecode, and runtime certificates/kubeconfigs. Tracked
fixtures and source files are preserved, as are `.cursor/`, `features/`, and
`context.md`. Regenerate the admin API before subsequent service builds.

The target does not prune Docker globally or clear shared Go caches. External
result directories, old temporary E2E snapshots, and installed tools are retained.
A failed runtime operation makes cleanup fail and retains local files for another
attempt. Repeat cleanup after resolving the error. Preview lists current actions;
networks that become empty during cleanup are removed on the actual run.
