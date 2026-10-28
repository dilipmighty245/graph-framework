# Nexus runtime packaging

This directory packages the infrastructure and installation resources used by a Nexus runtime. The core storage path consists of a Kubernetes API server backed by etcd; other components provide API access and model-specific behavior.

## Deployment paths

| Directory | Purpose |
| --- | --- |
| `helm-charts/core` | Core runtime resources. |
| `helm-charts/datamodel-init` | Datamodel installation resources. |
| `kind` | Local Kind environment scripts and configuration. |
| `k0s` | Local k0s environment scripts and configuration. |
| `k8s` | Standalone API-server and etcd Compose environment. |
| `Dockerfiles` | Supporting kubectl, etcd, and diagnostic images. |
| `fileUploadServer` | HTTP/WebDAV artifact service used by deployment workflows. |

The repository also distributes runtime files through datamodel templates and component charts. These paths are not interchangeable without reviewing their configuration.

## Before deployment

Choose the deployment path, configure registries and image versions, generate and install the admin model, and provision the required access and certificates. Inspect scripts before execution: install and uninstall targets create or delete infrastructure and persisted resources.

Several defaults refer to environment-specific registries and local development credentials. A production deployment needs its own credentials, storage policy, network configuration, and component configuration.

Framework build tags are local by default. Supply `DOCKER_REGISTRY` explicitly
for registry downloads or publication. The legacy k0s proxy requires
`NEXUS_KUBECTL_IMAGE`; the standalone API-server Compose workflow defaults to
`registry.k8s.io/kube-apiserver:v1.34.3`, overridable with
`NEXUS_KUBE_APISERVER_IMAGE`. Use images available through your environment's
registry policy. These workflows do not have a prepublished framework registry.

## Streaming compatibility

Generated SDKs and framework client modules use Kubernetes v0.34.3. Both core
charts and the Compose API server now select v1.34.3 with etcd v3.5.24. The modern
chart enables WatchList for its gateway with `watchListClient: true`. Other
applications enable it with `KUBE_FEATURE_WatchListClient=true` before startup;
client-go retains its LIST/WATCH fallback.

The legacy `core` chart uses TLS on port 6443 and an authenticated kubectl proxy
on port 8080. Existing callers keep `http://nexus-apiserver:8080` and
`http://nexus-proxy-container:80`. The controller-manager uses a TLS kubeconfig.
This preserves the existing trusted in-cluster HTTP access model; those proxy
services still provide administrative access and belong on a trusted network.

Both charts generate a fresh CA, API-server certificate, client certificate,
service-account signing key, and kubeconfigs in one Secret. Embedded certificates
and duplicate certificate jobs are removed. Helm upgrades reuse the Secret using
[lookup](https://helm.sh/docs/chart_template_guide/functions_and_pipelines/#using-the-lookup-function).
Credentials last `runtimeTLS.validityDays` (default 365); renewal requires
replacing the bundle and restarting its consumers. For GitOps/offline rendering,
where lookup cannot read the cluster, provision a stable Secret and set
`runtimeTLS.existingSecret=true` and `runtimeTLS.secretName`. Its required keys
are `ca.crt`, `apiserver.crt`, `apiserver.key`, `sa.key`, `kubeconfig` (service TLS
URL), and `proxy-kubeconfig` (TLS URL `https://127.0.0.1:6443`). Include the service
DNS names and loopback IP in the server certificate. Existing external TLS
clients must use the new CA/client kubeconfig when migrating from old certificates.

For registry-restricted environments, set these full image references to an
approved mirror:

| Deployment path | Overrides |
| --- | --- |
| Legacy `core` chart | `runtimeImages.apiserver`, `runtimeImages.controllerManager`, `runtimeImages.kubectl` |
| Modern `helm-charts/core` | `nexusKubeApiserver.image`, `runtimeImages.kubectl` |
| Compose | `NEXUS_KUBE_APISERVER_IMAGE`, `NEXUS_KUBECTL_IMAGE` |

For local testing with an already cached `kindest/node:v1.34.3`:

```sh
bash scripts/prepare-runtime-images.sh
# Use runtimeImages.kubectl=nexus/runtime-kubectl:v1.34.3 for Helm,
# or NEXUS_KUBECTL_IMAGE=nexus/runtime-kubectl:v1.34.3 for Compose.
```

The script imports the matching API-server/controller images and builds a
kubectl image from that node's binary without pulling images or changing registry
policy. It removes its disposable cluster and retains the local Docker images.
These images target the local Docker architecture. Loading them into another
cluster or publishing them to an approved mirror is a separate deployment step.
The modern chart exposes `global.etcd.storageClassName` (default `gp3-ebs`);
select a class present in your cluster. Other framework/helper images and etcd still need to be available through the
deployment's registry policy.

These source changes do not upgrade a running runtime. The disposable tests
exercise fresh storage; migration of a populated v1.19 runtime requires its own
backup and data-migration rehearsal before applying the new chart.

## Validation

Render Helm templates with representative values, check resource names and selectors, and run installation against a disposable environment. Verify API-server access, gateway readiness, model installation, generated-client operations, and cleanup.

See the [framework overview](../README.md), [admin model](../api/README.md), and [gateway guide](../api-gw/README.md) for component responsibilities.
