#!/usr/bin/env bash
# Reconnect the standalone server to the final SDK module after import rewriting.
set -euo pipefail
output=$(cd "$1" && pwd -P)
module_path=${2%/}
[[ -n "$module_path" ]] || { echo 'CRD_MODULE_PATH is required to finalize GraphQL.' >&2; exit 2; }
parent_module=$(cd "$output" && go list -mod=readonly -m -f '{{.Path}}')
parent_dir=$(cd "$output" && go list -mod=readonly -m -f '{{.Dir}}')
parent_dir=$(cd "$parent_dir" && pwd -P)
replacement=..
cursor=$output
while [[ "$cursor" != "$parent_dir" ]]; do
    [[ "$cursor" != / ]] || { echo 'SDK output is outside its module.' >&2; exit 2; }
    replacement="../$replacement"
    cursor=$(dirname "$cursor")
done
cd "$output/nexus-gql"
go mod edit "-module=$module_path/nexus-gql" \
    -droprequire=nexustempmodule -dropreplace=nexustempmodule \
    "-require=$parent_module@v0.0.0" "-replace=$parent_module=$replacement"
go mod tidy
