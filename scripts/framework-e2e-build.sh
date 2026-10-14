#!/usr/bin/env bash
# Runs inside the compiler image against an isolated working-tree snapshot.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
for tool in go git make jq yq goimports; do
    command -v "$tool" >/dev/null || { echo "Compiler image is missing $tool" >&2; exit 2; }
done
mkdir -p "$root/artifacts/e2e-bin"
export PATH="$root/artifacts/e2e-bin:$PATH"
# Build forked generators from the snapshot, including when using a builder image.
(cd "$root/gqlgen" && go build -o "$root/artifacts/e2e-bin/gqlgen" .)
(cd "$root/kube-openapi" && go build -o "$root/artifacts/e2e-bin/nexus-openapi-gen" ./cmd/nexus-openapi-gen)
cd "$root/compiler"
export DEFAULT_CLIENT_NAME
DEFAULT_CLIENT_NAME=$(yq eval '.k8s_clients.default' manifest.yaml)
codegen_tag=$(yq eval '.k8s_clients.versioned[] | select(.name == strenv(DEFAULT_CLIENT_NAME)) | .k8s_code_generator_git_tag' manifest.yaml)
[[ -n "$codegen_tag" && "$codegen_tag" != null ]] || { echo 'Missing code-generator version in manifest.yaml' >&2; exit 2; }
# Populate this dependency only inside the snapshot, never in the user's checkout.
mkdir -p _deps/github.com/kubernetes
if [[ ! -d _deps/github.com/kubernetes/code-generator/.git ]]; then
    git clone --depth 1 --branch "$codegen_tag" https://github.com/kubernetes/code-generator.git _deps/github.com/kubernetes/code-generator
fi
make generate_example COMPILER_SRC_DIRECTORY="$root/compiler"
cp "$root/compiler/_generated/go.mod" "$root/compiler/example/output/generated/go.mod"
cp "$root/compiler/_generated/go.sum" "$root/compiler/example/output/generated/go.sum"
cd "$root/compiler/example/output/generated"
go mod edit -module=github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated
go mod tidy
go build ./...
link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
go test -timeout=2m -count=1 -race -ldflags="$link_flags" ./nexus-client
cd "$root/compiler/example/tests"
# Replace directives of dependencies do not propagate to this main module.
for module in nexus gqlgen kube-openapi common-library; do
    go mod edit "-replace=github.com/dilipmighty245/graph-framework-for-microservices/$module=$root/$module"
done
go mod tidy
link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
go test -timeout=3m -count=1 -race -ldflags="$link_flags" ./...
mkdir -p "$root/artifacts/e2e-bin"
if [[ ,$E2E_COMPONENTS, == *,graphql,* ]]; then
    cd "$root/compiler/example/output/generated/nexus-gql"
    go mod edit "-replace=github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated=$root/compiler/example/output/generated"
    go mod tidy
    CGO_ENABLED=0 GOOS=linux GOARCH="$E2E_GOARCH" go build -o "$root/artifacts/e2e-bin/graphql" .
fi
if [[ ,$E2E_COMPONENTS, == *,gateway,* ]]; then
    service_test_args=(-timeout=3m -mod=readonly -count=1 -ldflags="$link_flags")
    if [[ ${FRAMEWORK_TEST_RACE:-false} == true ]]; then service_test_args+=(-race); fi
    cd "$root/compiler"
    rm -rf _generated
    mkdir -p "$root/api/build"
    make generate_code DATAMODEL_PATH=../api CONFIG_FILE=../api/nexus.yaml \
        GENERATED_OUTPUT_DIRECTORY=../api/build CRD_MODULE_PATH=nexus/admin/api/build/ \
        COMPILER_SRC_DIRECTORY="$root/compiler"
    cd "$root/api-gw"
    go mod tidy
    go test "${service_test_args[@]}" ./pkg/... ./internal/... ./cmd/...
    CGO_ENABLED=0 GOOS=linux GOARCH="$E2E_GOARCH" go build -o "$root/artifacts/e2e-bin/gateway" ./cmd/api-gw
    for module in authz-controller connector-controller; do
        cd "$root/$module"
        go test "${service_test_args[@]}" ./...
    done
fi
