# Nexus API gateway

The gateway exposes APIs for Nexus datamodels and routes requests to runtime resources and application services. It watches model and routing resources, assembles OpenAPI descriptions, and configures Envoy discovery resources.

## Responsibilities

- Track installed datamodels and CRDs.
- Serve generated and declarative API descriptions.
- Handle configured routes, proxy rules, and CORS behavior.
- Integrate OIDC authentication and tenant registration.
- Support admin and tenant gateway modes.

Some routes, registration hooks, and SKU configuration reflect product-specific integrations. Configure them for the deployment using the gateway.

## Building

The gateway imports generated types and clients from `../api/build`. Generate the [admin datamodel](../api/README.md) before building it; those packages are absent from a fresh checkout.

Once generated dependencies are available, build from this directory:

```sh
go build ./cmd/api-gw
go build ./cmd/openapi-generator
```

## Configuration and deployment

`KUBECONFIG` selects the runtime API. `APIGWCONFIG` can override the default configuration path `/config/api-gw-config`. See `pkg/config`, `deploy/config`, and command-line flags in `cmd/api-gw/main.go` for configuration behavior.

Deployment sources exist under `manifests`, `api-gateway`, and `deployment`. They represent different packaging paths and have different names, configuration, and certificate assumptions. Choose the path used by your installer and render it before applying it.

## Validation

Package tests cover routing, controllers, configuration, and API generation. Controller suites use Kubernetes envtest and expect local API-server, etcd, and kubectl binaries under `test/bin`. Generate the admin API and provision these assets before running the complete suite.

See the [framework overview](../README.md) for the gateway's role in the runtime.
