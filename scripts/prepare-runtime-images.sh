#!/usr/bin/env bash
# Import matching control-plane images from a cached Kind node, without a pull.
set -euo pipefail
version=v1.34.3
node_image="kindest/node:$version"
proxy_image=${NEXUS_KUBECTL_IMAGE:-nexus/runtime-kubectl:$version}
work=$(mktemp -d "${TMPDIR:-/tmp}/nexus-runtime-images-XXXXXXXX")
cluster=${RUNTIME_IMAGE_CLUSTER:-nexus-images-$(basename "$work" | tr '[:upper:]' '[:lower:]')}
owned=false
cleanup() {
    status=$?
    trap - EXIT
    if [[ "$owned" == true ]]; then
        kind delete cluster --name "$cluster" || status=1
    fi
    rm -rf "$work"
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if [[ -z ${RUNTIME_IMAGE_CLUSTER:-} ]]; then
    docker image inspect "$node_image" >/dev/null
    owned=true
    kind create cluster --name "$cluster" --image "$node_image" --kubeconfig "$work/kubeconfig" --wait 90s
fi
node=$(kind get nodes --name "$cluster" | head -1)
[[ -n "$node" ]] || { echo "No Kind node found for $cluster" >&2; exit 1; }
arch=$(docker version --format '{{.Server.Arch}}')
for component in kube-apiserver kube-controller-manager; do
    image="registry.k8s.io/$component:$version"
    # Export to stdout: no writes or registry requests inside the existing node.
    docker exec "$node" ctr --namespace k8s.io images export --platform "linux/$arch" - "$image" > "$work/$component.tar"
    docker load --input "$work/$component.tar"
done
kubectl_path=$(docker exec "$node" sh -c 'command -v kubectl')
docker cp "$node:$kubectl_path" "$work/kubectl"
cat > "$work/Dockerfile" <<'DOCKERFILE'
FROM scratch
COPY --chmod=755 kubectl /usr/local/bin/kubectl
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/kubectl"]
DOCKERFILE
docker build --pull=false --provenance=false --sbom=false --tag "$proxy_image" "$work"
docker run --rm --pull=never "$proxy_image" version --client
printf 'Images loaded locally. Set runtimeImages.kubectl=%s when testing the charts.\n' "$proxy_image"
