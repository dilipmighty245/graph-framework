#!/usr/bin/env python3
"""Remove local framework resources without deleting source or global caches."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


class Cleaner:
    def __init__(self, root, dry_run=False, cluster="", all_clusters=False):
        self.root = root
        self.dry_run = dry_run
        self.cluster = cluster
        self.all_clusters = all_clusters
        self.failed = False

    def query(self, *args):
        return subprocess.check_output(args, text=True).strip()

    def action(self, *args):
        print("+ " + " ".join(args), flush=True)
        if not self.dry_run:
            try:
                subprocess.run(args, check=True)
            except subprocess.CalledProcessError:
                self.failed = True

    def clusters(self, available):
        local = self.root / "nexus-runtime-manifests/kind"
        names = {p.name[1:] for p in local.glob(".*") if p.is_dir() and not p.is_symlink()}
        names.update(n for n in available if n.startswith("nexus-e2e-"))
        if self.cluster:
            names.add(self.cluster)
        if self.all_clusters:
            names.update(available)
        return names

    def docker(self):
        # No daemon/tool is a partial cleanup, never a silent success.
        if not shutil.which("docker"):
            print("Docker unavailable; runtime cleanup incomplete.", file=sys.stderr)
            self.failed = True
            return
        try:
            self.query("docker", "info")
            ids = self.query("docker", "ps", "-aq").split()
            containers = json.loads(self.query("docker", "inspect", *ids)) if ids else []
            available = sorted({(c.get("Config", {}).get("Labels") or {}).get("io.x-k8s.kind.cluster")
                                for c in containers} - {None, ""})
            clusters = self.clusters(available)
            if set(available) & clusters and not shutil.which("kind"):
                print("Kind unavailable; cluster cleanup incomplete.", file=sys.stderr)
                self.failed = True
            elif shutil.which("kind"):
                for name in sorted(set(available) & clusters):
                    self.action("kind", "delete", "cluster", "--name", name)
            compose_dirs = {str(self.root / "nexus-runtime-manifests" / n) for n in ("k0s", "k8s")}
            for c in containers:
                labels = c.get("Config", {}).get("Labels", {}) or {}
                name = c["Name"].lstrip("/")
                owned = (name == "nexus-api-gw" or
                         any(name in ("k8s-proxy-" + n, "nexus-api-gw-" + n) for n in clusters) or
                         labels.get("com.docker.compose.project.working_dir") in compose_dirs)
                if owned:
                    self.action("docker", "rm", "-fv", c["Id"])
            # Remove named volumes left behind by the two legacy Compose projects.
            volumes = self.query("docker", "volume", "ls", "-q").split()
            for volume in volumes:
                info = json.loads(self.query("docker", "volume", "inspect", volume))[0]
                labels = info.get("Labels") or {}
                if (labels.get("com.docker.compose.project") == "k0s" and
                        labels.get("com.docker.compose.volume") == "kubeconfig"):
                    self.action("docker", "volume", "rm", volume)
            tags = self.query("docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}}").splitlines()
            for tag in sorted(set(tags)):
                repo = tag.rsplit(":", 1)[0]
                if repo.startswith(("nexus/", "nexus-e2e/")) or repo in ("authz", "authz-builder"):
                    self.action("docker", "image", "rm", tag)
            networks = self.query("docker", "network", "ls", "--format", "{{.Name}}").splitlines()
            for name in ("nexus", "kind"):
                if name in networks:
                    info = json.loads(self.query("docker", "network", "inspect", name))[0]
                    # In preview, nodes/containers scheduled above are still attached.
                    if not info.get("Containers"):
                        self.action("docker", "network", "rm", name)
        except (subprocess.CalledProcessError, ValueError) as exc:
            print(f"Runtime cleanup incomplete: {exc}", file=sys.stderr)
            self.failed = True

    def files(self):
        # Select known output locations, then subtract tracked files. Never use git clean.
        patterns = (
            "**/artifacts", "api/build", "**/bin", "compiler/_generated",
            "compiler/_tsm_temp", "compiler/compiler.tar", "compiler/generated",
            "compiler/example/output",
            "compiler/example/datamodel/build",
            "compiler/example/test-utils/output-group-name-with-hyphen-datamodel",
            "compiler/cmd/nexus-sdk/nexus-sdk", "compiler/cmd/nexus-openapi-gen",
            "compiler/cmd/gqlgen", "compiler/cmd/validate", "api-gw/api-gw",
            "api-gw/cmd/openapi-generator/test/_generated",
            "authz-controller/cmd/authz-controller/authz-controller",
            "connector-controller/nexus",
            "nexus-runtime-manifests/kind/.*", "nexus-runtime-manifests/k0s/.kubeconfig",
            "nexus-runtime-manifests/k8s/certs", "nexus-runtime-manifests/k8s/ca*",
            "nexus-runtime-manifests/k8s/kubeconfig*", "**/__pycache__",
            "**/*.pyc", "**/*.test", "**/*.coverprofile", "**/coverage*.out",
        )
        tracked = set(subprocess.check_output(
            ["git", "-C", str(self.root), "ls-files", "-z"]).decode().split("\0"))
        protected = {".git", ".cursor", "features", ".agents", ".codex", ".aws"}

        def remove(path):
            rel = path.relative_to(self.root)
            if any(p in protected for p in rel.parts) or str(rel) in tracked:
                return
            # Do not traverse directory symlinks, including symlinked ancestors.
            if any(p.is_symlink() for p in path.parents if p != self.root and self.root in p.parents):
                return
            if path.is_dir() and not path.is_symlink():
                for child in path.iterdir():
                    remove(child)
                if not self.dry_run and not any(path.iterdir()):
                    path.rmdir()
            else:
                print(f"+ remove {rel}")
                if not self.dry_run:
                    path.unlink(missing_ok=True)

        for pattern in patterns:
            for path in self.root.glob(pattern):
                remove(path)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--all-kind-clusters", action="store_true")
    args = parser.parse_args()
    cluster = os.environ.get("CLUSTER_NAME", "")
    if cluster and not re.fullmatch(r"[a-z0-9][a-z0-9-]*", cluster):
        parser.error("CLUSTER_NAME must contain lowercase letters, digits, and hyphens")
    cleaner = Cleaner(ROOT, args.dry_run, cluster, args.all_kind_clusters)
    cleaner.docker()
    # Keep runtime configuration on failure so the next cleanup can identify it.
    if cleaner.failed:
        print("Runtime cleanup failed; local files retained for another attempt.", file=sys.stderr)
    else:
        cleaner.files()
    return int(cleaner.failed)


if __name__ == "__main__":
    sys.exit(main())
