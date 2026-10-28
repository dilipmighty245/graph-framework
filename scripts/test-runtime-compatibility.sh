#!/usr/bin/env bash
# Exercise only chart API infrastructure in a disposable Kind cluster.
# Requires a generated local example SDK; see docs/testing.md for preparation.
set -euo pipefail
umask 077
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
for tool in docker kind helm kubectl python3 go rg; do command -v "$tool" >/dev/null; done
python3 -c 'import yaml'
[[ -f "$root/compiler/example/output/generated/go.mod" ]] || {
    echo 'Generate the local example SDK before running this test.' >&2; exit 2;
}
work=$(mktemp -d "${TMPDIR:-/tmp}/nexus-runtime-test-XXXXXXXX")
cluster=nexus-runtime-$(basename "$work" | tr '[:upper:]' '[:lower:]')
results=${FRAMEWORK_TEST_RESULTS:-$root/artifacts/framework-tests}
mkdir -p "$results"
results=$(mktemp -d "$results/runtime-XXXXXXXX")
forward_pid=''
owned=false
kube() { kubectl --kubeconfig "$work/host-kubeconfig" "$@"; }
cleanup() {
    status=$?
    trap - EXIT
    set +e
    if [[ -n "$forward_pid" ]]; then kill "$forward_pid"; wait "$forward_pid" 2>/dev/null; fi
    if [[ "$owned" == true ]]; then
        kube get pods -A -o wide > "$results/pods.log" 2>&1
        for ns in runtime-legacy runtime-modern; do
            kube -n "$ns" get events > "$results/$ns-events.log" 2>&1
            kube -n "$ns" get deployments,statefulsets -o name | while read -r deployment; do
                kube -n "$ns" logs "$deployment" --all-containers > "$results/$ns-${deployment##*/}.log" 2>&1
            done
        done
        kind delete cluster --name "$cluster" > "$results/cleanup.log" 2>&1 || status=1
    fi
    rm -rf "$work"
    printf 'Runtime test logs: %s\n' "$results"
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
docker image inspect kindest/node:v1.34.3 >/dev/null
docker image inspect quay.io/coreos/etcd:v3.5.24 >/dev/null
owned=true
kind create cluster --name "$cluster" --image kindest/node:v1.34.3 --kubeconfig "$work/host-kubeconfig" --wait 90s > "$results/kind.log" 2>&1
RUNTIME_IMAGE_CLUSTER="$cluster" bash "$root/scripts/prepare-runtime-images.sh" > "$results/images.log" 2>&1
proxy_image=${NEXUS_KUBECTL_IMAGE:-nexus/runtime-kubectl:v1.34.3}
# API-server/controller images are already in this matching node. Export only
# this host platform: Docker's multi-platform cache can contain partial indexes
# that containerd's --all-platforms importer cannot load.
arch=$(docker version --format '{{.Server.Arch}}')
for image in quay.io/coreos/etcd:v3.5.24 "$proxy_image"; do
    docker image save --platform "linux/$arch" "$image" > "$work/image.tar"
    kind load image-archive --name "$cluster" "$work/image.tar" >> "$results/load.log" 2>&1
done
for chart_type in legacy modern; do
    ns=runtime-$chart_type
    chart="$root/nexus-runtime-manifests/core"
    api=nexus-apiserver
    proxy=nexus-proxy-container
    tls_template=templates/nexus-k8s-api.yaml
    if [[ "$chart_type" == modern ]]; then
        chart="$root/nexus-runtime-manifests/helm-charts/core"
        api=nexus-k8s-apiserver
        proxy=nexus-k8s-proxy
        tls_template=templates/runtime-tls.yaml
    fi
    kube create namespace "$ns"
    helm_args=(compat "$chart" --namespace "$ns" --set global.tainted=false --set global.resources.clustertype=dev --set global.resources.etcd.cpu=250m --set global.etcd.storageClassName=standard --set "runtimeImages.kubectl=$proxy_image")
    # Rendered files contain private keys. Keep them only in the private temp dir.
    helm template "${helm_args[@]}" > "$work/render.yaml"
    python3 - "$work/render.yaml" "$work/infra.yaml" "$ns" <<'PY'
import sys, yaml
documents = [d for d in yaml.safe_load_all(open(sys.argv[1])) if d]
seen = set()
selected = []
for d in documents:
    kind, name = d['kind'], d['metadata']['name']
    key = (kind, name)
    assert key not in seen, f'Duplicate chart resource: {key}'
    seen.add(key)
    if kind == 'Secret' or (kind in ('Deployment', 'StatefulSet', 'Service') and name in (
            'nexus-etcd', 'nexus-etcd-headless', 'nexus-kube-apiserver',
            'nexus-kube-controllermanager', 'nexus-apiserver', 'nexus-proxy-container',
            'nexus-k8s-apiserver', 'nexus-k8s-proxy')):
        # Match metadata Helm adds during a real install, so its server-side
        # dry run can check lookup/reuse against this infrastructure-only install.
        d['metadata']['labels'] = d['metadata'].get('labels') or {}
        d['metadata']['annotations'] = d['metadata'].get('annotations') or {}
        d['metadata']['labels']['app.kubernetes.io/managed-by'] = 'Helm'
        d['metadata']['annotations'].update({
            'meta.helm.sh/release-name': 'compat',
            'meta.helm.sh/release-namespace': sys.argv[3],
        })
        selected.append(d)
with open(sys.argv[2], 'w') as output:
    yaml.safe_dump_all(selected, output)
PY
    kube apply -f "$work/infra.yaml" > "$results/$ns-apply.log"
    kube -n "$ns" rollout status statefulset/nexus-etcd --timeout=120s
    kube -n "$ns" rollout status deployment/$( [[ "$chart_type" == legacy ]] && echo nexus-kube-apiserver || echo nexus-k8s-apiserver ) --timeout=180s
    if [[ "$chart_type" == legacy ]]; then
        kube -n "$ns" rollout status deployment/nexus-kube-controllermanager --timeout=120s
    else
        kube -n "$ns" rollout status deployment/nexus-k8s-proxy --timeout=120s
    fi
    # Verify Helm's server-side lookup preserves certificates on upgrades.
    helm template "${helm_args[@]}" --kubeconfig "$work/host-kubeconfig" --dry-run=server --show-only "$tls_template" > "$work/reused.yaml"
    python3 - "$work/infra.yaml" "$work/reused.yaml" <<'PY'
import sys, yaml
def secrets(path):
    return {d['metadata']['name']: d['data'] for d in yaml.safe_load_all(open(path)) if d and d['kind'] == 'Secret'}
assert secrets(sys.argv[1]) == secrets(sys.argv[2]), 'TLS credentials changed on upgrade'
PY
    kube -n "$ns" port-forward "service/$proxy" :80 > "$work/forward.log" 2>&1 &
    forward_pid=$!
    for ((attempt=0; attempt<100; attempt++)); do
        if rg -q 'Forwarding from 127.0.0.1:' "$work/forward.log"; then break; fi
        kill -0 "$forward_pid"
        sleep .1
    done
    port=$(sed -n 's/Forwarding from 127.0.0.1:\([0-9]*\).*/\1/p' "$work/forward.log" | head -1)
    [[ -n "$port" ]]
    cat > "$work/nested-kubeconfig" <<EOF
apiVersion: v1
kind: Config
clusters:
- name: runtime
  cluster:
    server: http://127.0.0.1:$port
users:
- name: runtime
contexts:
- name: runtime
  context:
    cluster: runtime
    user: runtime
current-context: runtime
EOF
    nested() { kubectl --kubeconfig "$work/nested-kubeconfig" "$@"; }
    for ((attempt=0; attempt<90; attempt++)); do
        if nested get --raw /readyz > /dev/null 2>&1; then break; fi
        sleep 1
    done
    nested get --raw /version > "$results/$ns-version.json"
    if [[ "$chart_type" == legacy ]]; then
        kube -n "$ns" exec deployment/nexus-kube-apiserver -c api-proxy -- \
            kubectl --server=http://nexus-apiserver:8080 get --raw /readyz > "$results/$ns-http-endpoint.log"
        nested create namespace controller-check
        for ((attempt=0; attempt<60; attempt++)); do
            if nested -n controller-check get serviceaccount default >/dev/null 2>&1; then break; fi
            sleep 1
        done
        nested -n controller-check get serviceaccount default > "$results/$ns-controller.log"
        nested delete namespace controller-check --wait=false
    fi
    nested apply -f "$root/compiler/example/output/generated/crds" > "$results/$ns-crds.log"
    nested wait --for=condition=Established --timeout=90s -f "$root/compiler/example/output/generated/crds" >> "$results/$ns-crds.log"
    link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
    if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
    for streaming in false true; do
        (cd "$root/compiler/example/tests" && KUBECONFIG="$work/nested-kubeconfig" \
            KUBE_FEATURE_WatchListClient="$streaming" SUBSCRIPTION_EXPECT_WATCH_LIST="$streaming" \
            go test -tags=integration -mod=readonly -race -count=1 -timeout=4m -v -ldflags="$link_flags" ./integration) \
            > "$results/$ns-subscriptions-$streaming.log" 2>&1
    done
    kill "$forward_pid"
    wait "$forward_pid" 2>/dev/null || true
    forward_pid=''
    printf '%s: startup, credential reuse, and both subscription modes passed\n' "$chart_type"
done
