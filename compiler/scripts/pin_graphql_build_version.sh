#!/usr/bin/env bash
set -euo pipefail

# The generated GraphQL binary and SDK must use the same Kubernetes libraries.
compiler_source=$1
bash "$compiler_source/scripts/pin_deps.sh" "$compiler_source"
go mod edit -require=github.com/vektah/gqlparser/v2@v2.5.0
