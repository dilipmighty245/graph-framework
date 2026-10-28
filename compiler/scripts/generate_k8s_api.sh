#!/usr/bin/env bash
set -euo pipefail

export DEFAULT_CLIENT_NAME
DEFAULT_CLIENT_NAME=$(yq eval '.k8s_clients.default' manifest.yaml)
version=$(yq eval '.k8s_clients.versioned[] | select(.name == strenv(DEFAULT_CLIENT_NAME)) | .k8s_code_generator_git_tag' manifest.yaml)
[[ -n "$version" && "$version" != null ]] || { echo 'Missing code-generator version' >&2; exit 1; }
# gengo resolves package directories physically; macOS TMPDIR can be a symlink.
codegen=$(cd _deps/github.com/kubernetes/code-generator && pwd -P)
git -C "$codegen" checkout "$version"
source "$codegen/kube_codegen.sh"

generated=$(cd _generated && pwd -P)
(
  cd "$generated"
  kube::codegen::gen_helpers \
    --boilerplate "$codegen/examples/hack/boilerplate.go.txt" \
    "$generated/apis"
  kube::codegen::gen_client \
    --with-watch \
    --output-dir "$generated/client" \
    --output-pkg nexustempmodule/client \
    --boilerplate "$codegen/examples/hack/boilerplate.go.txt" \
    "$generated/apis"
)
mkdir -p "$GOPATH/src/nexustempmodule"
