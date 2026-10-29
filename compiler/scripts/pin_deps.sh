#!/usr/bin/env bash
set -euo pipefail

compiler_source=$1
export DEFAULT_CLIENT_NAME
DEFAULT_CLIENT_NAME=$(yq eval '.k8s_clients.default' "$compiler_source/manifest.yaml")
version=$(yq eval '.k8s_clients.versioned[] | select(.name == strenv(DEFAULT_CLIENT_NAME)) | .k8s_code_generator_git_tag' "$compiler_source/manifest.yaml")
[[ -n "$version" && "$version" != null ]] || { echo 'Missing Kubernetes client version' >&2; exit 1; }

go mod edit -go=1.27.3
# Keep SDK runtime dependencies explicit. Transitive versions come from these
# modules and the copied datamodel go.mod; unused controller/cert-manager pins
# would impose obsolete constraints on SDKs that do not import those packages.
for module in api apimachinery client-go apiextensions-apiserver; do
  go mod edit "-require=k8s.io/$module@$version" "-replace=k8s.io/$module=k8s.io/$module@$version"
done
for module in gqlgen kube-openapi nexus; do
  go mod edit "-replace=github.com/dilipmighty245/graph-framework-for-microservices/$module=$compiler_source/../$module"
done
go mod edit -require=github.com/elliotchance/orderedmap@v1.4.0 -require=github.com/sirupsen/logrus@v1.9.3
