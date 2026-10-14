#!/usr/bin/env bash
# Repeatable checks across module boundaries. No module files are rewritten.
set -euo pipefail
export PYTHONDONTWRITEBYTECODE=1
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
mode=${1:-regression}
case "$mode" in regression|compiler|gateway|all) ;; *) echo "Usage: $0 [regression|compiler|gateway|all]" >&2; exit 2 ;; esac
command -v go >/dev/null || { echo 'Go is required.' >&2; exit 2; }
results=${FRAMEWORK_TEST_RESULTS:-"$root/artifacts/framework-tests"}
mkdir -p "$results"
results=$(mktemp -d "$results/run-XXXXXXXX")
printf 'Results: %s\n' "$results"
go version | tee "$results/toolchain.log"
# Older Go linkers need the external linker on recent macOS versions.
link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
failures=0
run_check() {
    local name=$1 module=$2
    shift 2
    printf '\nChecking %s (%s)\n' "$name" "$module"
    if (cd "$root/$module" && "$@") 2>&1 | tee "$results/$name.log"; then
        printf 'PASS\t%s\n' "$name" >> "$results/summary.tsv"
    else
        printf 'FAIL\t%s\n' "$name" >> "$results/summary.tsv"
        failures=$((failures + 1))
    fi
}
test_args=(-timeout=5m -mod=readonly -count=1 -shuffle=on -ldflags="$link_flags")
if [[ ${FRAMEWORK_TEST_RACE:-false} == true ]]; then test_args+=(-race); fi
runs=${FRAMEWORK_TEST_RUNS:-1}
[[ $runs =~ ^[1-9][0-9]*$ ]] || { echo 'FRAMEWORK_TEST_RUNS must be a positive integer.' >&2; exit 2; }
for ((run=1; run<=runs; run++)); do
    printf '\nPass %s/%s\n' "$run" "$runs"
    if [[ $mode == regression || $mode == all ]]; then
        run_check "source-hygiene-$run" . python3 scripts/check-repository-hygiene.py
        run_check "runner-$run" . python3 -m unittest discover -s scripts/tests -v
        for module in nexus common-library api validation install-validator cli calibration connector graphql-server; do
            run_check "$module-$run" "$module" go test "${test_args[@]}" ./...
        done
        run_check "gateway-modules-$run" . bash scripts/check-gateway.sh modules
        run_check "openapi-$run" kube-openapi go test "${test_args[@]}" ./...
        run_check "graphql-generator-$run" gqlgen go test "${test_args[@]}" ./...
        run_check "kubebuilder-$run" kubebuilder go test "${test_args[@]}" ./pkg/... ./cmd/... ./test/e2e/utils/...
        run_check "file-upload-$run" nexus-runtime-manifests/fileUploadServer go test "${test_args[@]}" ./...
    fi
    if [[ $mode == compiler || $mode == regression || $mode == all ]]; then
        run_check "compiler-$run" compiler go test "${test_args[@]}" ./pkg/... ./cmd/...
    fi
    if [[ $mode == gateway ]]; then
        run_check "gateway-compile-$run" . bash scripts/check-gateway.sh
        run_check "gateway-unit-$run" api-gw go test "${test_args[@]}" ./pkg/... ./internal/... ./cmd/...
        for module in authz-controller connector-controller; do
            run_check "$module-$run" "$module" go test "${test_args[@]}" ./...
        done
    fi
    if [[ $mode == all ]]; then
        run_check "kind-e2e-$run" . bash scripts/test-framework-e2e.sh
    fi
done
printf '\n'
cat "$results/summary.tsv"
printf '\nLogs: %s\n' "$results"
if (( failures != 0 )); then
    printf '%s check(s) failed. See the logs; failures are not skipped.\n' "$failures" >&2
    exit 1
fi
