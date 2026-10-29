#!/usr/bin/env bash
# Build public Go builders and run generation/services/events from their sources.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
for tool in docker go python3; do
    command -v "$tool" >/dev/null || { echo "Builder test prerequisite missing: $tool" >&2; exit 2; }
done
results=${FRAMEWORK_TEST_RESULTS:-"$root/artifacts/framework-tests"}
mkdir -p "$results"
results=$(mktemp -d "$results/builders-XXXXXXXX")
printf 'Builder results: %s\n' "$results"
export E2E_COMPILER_IMAGE=${COMPILER_BUILDER_IMAGE:-nexus/compiler-builder:go1.27.3}
export E2E_AUTHZ_BUILDER_IMAGE=${AUTHZ_BUILDER_IMAGE:-nexus/authz-builder:go1.27.3}
build_args=(--progress=plain)
for component in compiler authz-controller; do
    image=$E2E_COMPILER_IMAGE
    if [[ $component == authz-controller ]]; then image=$E2E_AUTHZ_BUILDER_IMAGE; fi
    docker build "${build_args[@]}" -t "$image" "$root/$component/builder" 2>&1 | tee "$results/$component-build.log"
    docker run --rm "$image" bash -euc \
        'go version | grep -E "go1[.]27[.]3 linux/"; goimports -h 2>&1 | head -1; golangci-lint version' \
        2>&1 | tee "$results/$component-tools.log"
done
export FRAMEWORK_TEST_RACE=${FRAMEWORK_TEST_RACE:-true}
export FRAMEWORK_TEST_RESULTS=$results
bash "$root/scripts/test-framework-e2e.sh"
printf 'PASS: both public builders, fresh generation, subscription events, and services\n' | tee "$results/summary.log"
