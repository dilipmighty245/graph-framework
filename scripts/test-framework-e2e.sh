#!/usr/bin/env bash
# Generate from current sources, then exercise resources/services in a new Kind cluster.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
export E2E_COMPONENTS=${E2E_COMPONENTS:-resources,gateway,graphql}
E2E_KIND_IMAGE=${E2E_KIND_IMAGE:-kindest/node:v1.34.3}
for component in ${E2E_COMPONENTS//,/ }; do
    case "$component" in resources|gateway|graphql) ;; *) echo "Unknown E2E component: $component" >&2; exit 2 ;; esac
done
[[ ,$E2E_COMPONENTS, == *,resources,* ]] || { echo 'E2E_COMPONENTS must include resources.' >&2; exit 2; }
E2E_COMPILER_IMAGE=${E2E_COMPILER_IMAGE:-}
for tool in docker kind kubectl git python3 go; do
    command -v "$tool" >/dev/null || { echo "E2E prerequisite missing: $tool" >&2; exit 2; }
done
docker info >/dev/null 2>&1 || { echo 'A running Docker daemon is required.' >&2; exit 2; }
if [[ -n "$E2E_COMPILER_IMAGE" ]]; then
 docker image inspect "$E2E_COMPILER_IMAGE" >/dev/null 2>&1 || { echo "Build the compiler image locally first: $E2E_COMPILER_IMAGE" >&2; exit 2; }
fi
export E2E_GOARCH
E2E_GOARCH=$(docker version --format '{{.Server.Arch}}')
export GOMODCACHE=${GOMODCACHE:-$(go env GOMODCACHE)}
results=${FRAMEWORK_TEST_RESULTS:-"$root/artifacts/framework-tests"}
mkdir -p "$results"
results=$(mktemp -d "$results/e2e-XXXXXXXX")
work=$(mktemp -d "${TMPDIR:-/tmp}/nexus-e2e-XXXXXXXX")
cluster="nexus-e2e-$(basename "$work" | tr '[:upper:]' '[:lower:]')"
cluster_created=false
forward_pids=""
created_images=""
cleanup() {
    status=$?
    trap - EXIT INT TERM
    set +e
    for pid in $forward_pids; do kill "$pid" 2>/dev/null; wait "$pid" 2>/dev/null; done
    if [[ "$cluster_created" == true ]]; then
        if (( status != 0 )); then
            kubectl --kubeconfig "$work/kubeconfig" get pods -A -o wide > "$results/pods.log" 2>&1
            kubectl --kubeconfig "$work/kubeconfig" get events -A > "$results/events.log" 2>&1
            for service in gateway graphql; do
                kubectl --kubeconfig "$work/kubeconfig" -n framework-e2e logs "deployment/$service" --all-containers > "$results/$service.log" 2>&1
            done
        fi
        if ! kind delete cluster --name "$cluster" > "$results/cleanup.log" 2>&1; then
            echo "Cluster cleanup failed; run: kind delete cluster --name $cluster" >&2
            status=1
        fi
    fi
    for image in $created_images; do docker image rm "$image" > /dev/null 2>&1; done
    rm -rf "$work"
    printf 'E2E logs: %s\n' "$results"
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
printf 'E2E components: %s\nCluster: %s\nLogs: %s\n' "$E2E_COMPONENTS" "$cluster" "$results"
# Copy current tracked and untracked source, including uncommitted edits. Avoid
# IDE plans and feature notes. Never reset generated files in the original tree.
python3 - "$root" "$work/repo" <<'PY'
import os, pathlib, shutil, subprocess, sys
source, target = map(pathlib.Path, sys.argv[1:])
target.mkdir()
files = subprocess.check_output(["git", "-C", str(source), "ls-files", "-z", "--cached", "--others", "--exclude-standard"])
for raw in set(files.split(b"\0")):
    if not raw:
        continue
    rel = pathlib.Path(os.fsdecode(raw))
    if rel.parts[0] in (".cursor", "features"):
        continue
    src, dst = source / rel, target / rel
    if src.is_file():
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)
PY
# Some compiler targets restore generated helper files with git checkout.
git -C "$work/repo" init -q
git -C "$work/repo" add -A
git -C "$work/repo" -c user.name=framework-e2e -c user.email=framework-e2e@localhost commit -qm 'Isolated E2E working-tree snapshot'
if [[ -n "$E2E_COMPILER_IMAGE" ]]; then
container_root=/go/src/github.com/dilipmighty245/graph-framework-for-microservices
docker_args=(--volume "$GOMODCACHE:/go/pkg/mod" --volume "${GOCACHE:-$(go env GOCACHE)}:/root/.cache/go-build")
docker run --rm --pull=never \
    "${docker_args[@]}" \
    --volume "$work/repo:$container_root" --workdir "$container_root" \
    -e E2E_COMPONENTS -e E2E_GOARCH -e FRAMEWORK_TEST_RACE --entrypoint /bin/bash "$E2E_COMPILER_IMAGE" \
    scripts/framework-e2e-build.sh 2>&1 | tee "$results/generation.log"
if [[ -n ${E2E_AUTHZ_BUILDER_IMAGE:-} ]]; then
    [[ ,$E2E_COMPONENTS, == *,gateway,* ]] || { echo 'Authorization builder tests need gateway/admin model generation.' >&2; exit 2; }
    docker run --rm --pull=never "${docker_args[@]}" \
        --volume "$work/repo:$container_root" --workdir "$container_root/authz-controller" \
        "$E2E_AUTHZ_BUILDER_IMAGE" go test -timeout=4m -mod=readonly -race -count=1 ./... \
        2>&1 | tee "$results/authorization-builder.log"
fi
# Generated module replacements must point to the mounted tree on this host
# when running the live subscription tests outside the Linux build container.
python3 - "$work/repo" "$container_root" <<'PY'
from pathlib import Path
import sys
root, container_root = Path(sys.argv[1]), sys.argv[2]
for path in root.rglob("go.mod"):
    content = path.read_text()
    updated = content.replace(container_root, str(root))
    if updated != content:
        path.write_text(updated)
PY
else
    for tool in go make jq yq goimports; do
        command -v "$tool" >/dev/null || { echo "Host build prerequisite missing: $tool" >&2; exit 2; }
    done
    mkdir -p "$work/bin" "$work/gopath"
    # Keep the user's module cache while isolating GOPATH generation outputs.
    export GOMODCACHE=${GOMODCACHE:-$(go env GOMODCACHE)}
    export GOPATH="$work/gopath"
    export PATH="$work/bin:$PATH"
    bash "$work/repo/scripts/framework-e2e-build.sh" 2>&1 | tee "$results/generation.log"
fi
kind_args=(create cluster --name "$cluster" --kubeconfig "$work/kubeconfig" --wait 120s)
if [[ -n ${E2E_KIND_IMAGE:-} ]]; then kind_args+=(--image "$E2E_KIND_IMAGE"); fi
# Mark ownership before creating, so cleanup also handles partial creation.
cluster_created=true
kind "${kind_args[@]}" 2>&1 | tee "$results/kind.log"
kube() { kubectl --kubeconfig "$work/kubeconfig" "$@"; }
crds="$work/repo/compiler/example/output/generated/crds"
kube apply -f "$crds" 2>&1 | tee "$results/crds.log"
kube wait --for=condition=Established --timeout=90s -f "$crds"
# Run against the freshly generated SDK and the real API server. Each pass
# creates and deletes its own graph and starts a fresh Go test process.
sdk_runs=${E2E_SDK_RUNS:-3}
[[ $sdk_runs =~ ^[1-9][0-9]*$ ]] || { echo 'E2E_SDK_RUNS must be a positive integer.' >&2; exit 2; }
link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
run_subscriptions() {
    local mode=$1 streaming=$2 sdk_run
    for ((sdk_run=1; sdk_run<=sdk_runs; sdk_run++)); do
        (cd "$work/repo/compiler/example/tests" && KUBECONFIG="$work/kubeconfig" \
            KUBE_FEATURE_WatchListClient="$streaming" SUBSCRIPTION_EXPECT_WATCH_LIST="$streaming" \
            go test -tags=integration -timeout=4m -mod=readonly -race -count=1 -v \
            -ldflags="$link_flags" ./integration) 2>&1 | tee "$results/subscriptions${mode}-$sdk_run.log"
    done
}
run_subscriptions '' false
if [[ ${E2E_SUBSCRIPTION_COMPARE_STREAMING:-true} == true ]]; then
    # Both modes use the production dependency graph and freshly generated code.
    (cd "$work/repo/compiler/example/tests" && go list -m k8s.io/api k8s.io/apimachinery k8s.io/client-go) \
        2>&1 | tee "$results/subscriptions-versions.log"
    run_subscriptions '-streaming' true
fi
# These names follow common-library/pkg/nexus's hierarchy-aware naming contract.
root_name=$(python3 -c 'import hashlib; print(hashlib.sha1(b"roots.root.nexus.com:default").hexdigest())')
config_name=$(python3 -c 'import hashlib; print(hashlib.sha1(b"roots.root.nexus.com:default/configs.config.nexus.com:default").hexdigest())')
python3 - "$work/objects.json" "$root_name" "$config_name" <<'PY'
import json, sys
path, root_name, config_name = sys.argv[1:]
def resource(kind, group, name, spec):
    return {"apiVersion": group + "/v1", "kind": kind,
            "metadata": {"name": name, "labels": {"nexus/display_name": "default", "nexus/is_name_hashed": "true"}}, "spec": spec}
root = resource("Root", "root.nexus.com", root_name, {})
config = resource("Config", "config.nexus.com", config_name, {
    "myStr0": "before", "myStr1": [], "myStr2": {}, "xYZPort": 8080,
    "aBCHost": [], "clusterNamespaces": [], "instance": 0, "option_cu": "",
    "testValMarkers": {"myStr": "ab", "myInt": 3, "mySlice": ["a", "b"]},
})
config["metadata"]["labels"]["roots.root.nexus.com"] = "default"
with open(path, "w") as f:
    json.dump({"apiVersion": "v1", "kind": "List", "items": [root, config]}, f)
PY
kube create -f "$work/objects.json"
if kube patch configs.config.nexus.com "$config_name" --type=merge \
    -p '{"spec":{"testValMarkers":{"mySlice":["a","a"]}}}' > "$results/validation.log" 2>&1; then
    echo 'CRD validation accepted duplicate values in a unique list.' >&2
    exit 1
fi
if ! grep -qi 'duplicate' "$results/validation.log"; then
    cat "$results/validation.log" >&2
    echo 'Unique-list validation failed for an unexpected reason.' >&2
    exit 1
fi
kube get configs.config.nexus.com "$config_name" -o json > "$results/created.json"
python3 - "$results/created.json" <<'PY'
import json, sys
with open(sys.argv[1]) as f: obj = json.load(f)
assert obj["spec"]["myStr0"] == "before", obj
assert obj["metadata"]["labels"]["roots.root.nexus.com"] == "default", obj
PY
kube patch configs.config.nexus.com "$config_name" --type=merge -p '{"spec":{"myStr0":"after"}}'
kube get configs.config.nexus.com "$config_name" -o json > "$results/updated.json"
python3 - "$results/updated.json" <<'PY'
import json, sys
with open(sys.argv[1]) as f: obj = json.load(f)
assert obj["spec"]["myStr0"] == "after", obj
PY
if [[ ,$E2E_COMPONENTS, == *,gateway,* || ,$E2E_COMPONENTS, == *,graphql,* ]]; then
    kube create namespace framework-e2e
    kube -n framework-e2e create serviceaccount framework-e2e
    kube create clusterrolebinding framework-e2e --clusterrole=cluster-admin --serviceaccount=framework-e2e:framework-e2e
fi
forward() {
    local service=$1 pid port='' attempt
    kubectl --kubeconfig "$work/kubeconfig" -n framework-e2e port-forward --address 127.0.0.1 "service/$service" :8080 > "$results/$service-forward.log" 2>&1 &
    pid=$!
    forward_pids="$forward_pids $pid"
    for ((attempt=0; attempt<100; attempt++)); do
        port=$(sed -n 's/^Forwarding from 127\.0\.0\.1:\([0-9]*\) -> .*/\1/p' "$results/$service-forward.log" | head -1)
        if [[ -n "$port" ]]; then forwarded_port=$port; return; fi
        kill -0 "$pid" 2>/dev/null || { cat "$results/$service-forward.log" >&2; return 1; }
        sleep 0.1
    done
    echo "Port forwarding timed out for $service" >&2
    return 1
}
if [[ ,$E2E_COMPONENTS, == *,gateway,* ]]; then
    link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
    if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
    (cd "$work/repo/api-gw" && USE_EXISTING_CLUSTER=true KUBECONFIG="$work/kubeconfig" \
        go test -timeout=4m -mod=readonly -count=1 -race -ldflags="$link_flags" ./controllers) \
        2>&1 | tee "$results/gateway-controllers.log"
fi
for component in ${E2E_COMPONENTS//,/ }; do
    [[ $component != resources ]] || continue
    if [[ $component == gateway ]]; then
        kube apply -f "$work/repo/api/build/crds"
        kube wait --for=condition=Established --timeout=90s -f "$work/repo/api/build/crds"
    fi
    image="nexus-e2e/$component:$(basename "$work" | tr '[:upper:]' '[:lower:]')"
    printf 'FROM scratch\nCOPY %s /server\nENTRYPOINT ["/server"]\n' "$component" > "$work/repo/Dockerfile.e2e"
    docker build --platform "linux/$E2E_GOARCH" -f "$work/repo/Dockerfile.e2e" -t "$image" "$work/repo/artifacts/e2e-bin" 2>&1 | tee "$results/$component-image.log"
    created_images="$created_images $image"
    kind load docker-image "$image" --name "$cluster"
    kube -n framework-e2e create deployment "$component" --image="$image"
    kube -n framework-e2e set serviceaccount "deployment/$component" framework-e2e
    kube -n framework-e2e set env "deployment/$component" PORT=8080 NAMESPACE=framework-e2e
    if [[ $component == gateway ]]; then
        cat > "$work/api-gw-config" <<'YAML'
server:
  address: ":8080"
  httpPort: "8080"
  healthProbeAddrress: ":8081"
  metricsAddress: ":8082"
  certPath: "/unused/cert.pem"
  keyPath: "/unused/key.pem"
enable_nexus_runtime: true
YAML
        printf 'Prefix: []\nSuffix: []\n' > "$work/staticRoutes.yaml"
        kube -n framework-e2e create configmap gateway-config --from-file="api-gw-config=$work/api-gw-config" --from-file="staticRoutes.yaml=$work/staticRoutes.yaml"
        kube -n framework-e2e patch deployment gateway --type=strategic -p '{"spec":{"template":{"spec":{"volumes":[{"name":"config","configMap":{"name":"gateway-config"}}],"containers":[{"name":"gateway","volumeMounts":[{"name":"config","mountPath":"/config"}],"readinessProbe":{"httpGet":{"path":"/readyz","port":8081},"periodSeconds":2}}]}}}}'
    fi
    kube -n framework-e2e rollout status "deployment/$component" --timeout=120s
    kube -n framework-e2e expose deployment "$component" --port=8080 --target-port=8080
    forward "$component"
    if [[ $component == gateway ]]; then
        path="/apis/config.nexus.com/v1/configs/$config_name"
    else
        path=/query
    fi
    python3 "$root/scripts/framework-http-probe.py" "$component" "http://127.0.0.1:$forwarded_port$path" --name "$config_name"
done
kube delete -f "$work/objects.json" --wait=true --timeout=60s
for resource in "configs.config.nexus.com/$config_name" "roots.root.nexus.com/$root_name"; do
    remaining=$(kube get "$resource" --ignore-not-found -o name)
    [[ -z "$remaining" ]] || { echo "Resource remained after deletion: $remaining" >&2; exit 1; }
done
if [[ ${E2E_RUNTIME_COMPATIBILITY:-false} == true ]]; then
    # Reuse this run's freshly generated SDK before its isolated snapshot is
    # deleted. Never depend on generated files in the user's working tree.
    FRAMEWORK_TEST_RESULTS="$results" bash "$work/repo/scripts/test-runtime-compatibility.sh" \
        2>&1 | tee "$results/runtime-compatibility.log"
fi
printf 'PASS: generation, generated-client tests, resource CRUD, and selected service probes\n' | tee "$results/summary.log"
