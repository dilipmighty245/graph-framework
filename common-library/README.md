# Shared Nexus helpers

This module contains naming and schema comparison functions shared by framework components.

## Packages

| Package | Purpose |
| --- | --- |
| `pkg/nexus` | Parses parent labels and computes hierarchy-aware resource names. |
| `pkg/nexus-compare` | Compares model definitions and produces reports used during installation and upgrade. |

Naming is a compatibility contract between clients, admission handling, and stored resources. Changing its inputs or hashing behavior can change object identity. Comparison rules also influence upgrade decisions and should be checked against existing model fixtures.

## Development

Run checks from this directory:

```sh
go test ./...
go vet ./...
```

Consumers use local replacements to this module. See [install-validator](../install-validator/README.md) and the [framework overview](../README.md) for context.
