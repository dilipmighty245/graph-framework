# Datamodel installation validator

This tool checks CRD changes during datamodel installation or upgrade. It compares proposed definitions with the installed model and applies the changes through Kubernetes clients.

## Inputs

The command at `cmd/validate` reads:

- `CRD_SPEC_DIR`: CRD directory; defaults to `/crds`.
- `FORCE`: enables the force option when set to `true`.
- `/NAME`: file identifying the datamodel group.
- Kubernetes client configuration for the target runtime.

Inspect the comparison report before choosing a forced upgrade. Schema compatibility and the effect on existing objects need application-level verification.

## Development and packaging

From this directory:

```sh
go test ./...
go build ./cmd/validate
```

The module uses the sibling [common-library](../common-library/README.md) for comparison logic. The compiler packages the executable as an installation artifact; its build uses the filename `install.so`, although it is built as an executable rather than with Go's plugin build mode.

See the [compiler guide](../compiler/README.md) for artifact generation.
