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

## Validation

Render Helm templates with representative values, check resource names and selectors, and run installation against a disposable environment. Verify API-server access, gateway readiness, model installation, generated-client operations, and cleanup.

See the [framework overview](../README.md), [admin model](../api/README.md), and [gateway guide](../api-gw/README.md) for component responsibilities.
