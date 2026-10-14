# Nexus validation webhook

The validation service handles admission requests for Nexus resources. It implements resource naming and validation or mutation behavior associated with generated datamodels.

## Role in the runtime

The generated client, resource metadata, and admission layer must agree on hierarchy and naming. Changes to these conventions should be checked together with the compiler and shared naming helpers.

Source is organized under `pkg/config`, `pkg/validate`, and `cmd/nexus-validation`. Deployment manifests and Helm templates configure the webhook service, certificates, and runtime connection.

## Development

From this directory:

```sh
go test ./...
go build ./cmd/nexus-validation
```

The module depends on the sibling `common-library` through a local replacement. Keep both directories available. Inspect `Makefile` before using its container and test targets; some targets force Linux execution and need an appropriate host or container.

## Deployment

Review `values.yaml`, `templates`, and `manifests` for image, certificate, namespace, and service configuration. Webhook availability and certificate trust must be established before sending admission traffic to it.

See the [framework overview](../README.md) for the resource lifecycle.
