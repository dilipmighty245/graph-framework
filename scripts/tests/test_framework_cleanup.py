"""Exercise destructive cleanup boundaries using temporary files and fake Docker."""
import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).resolve().parents[1] / "clean-framework.py"
spec = importlib.util.spec_from_file_location("framework_clean", SCRIPT)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class CleanupTests(unittest.TestCase):
    def test_files_preserve_tracked_notes_and_symlink_targets_and_repeat(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "repo"
            root.mkdir()
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            tracked = root / "api/build/fixture.go"
            tracked.parent.mkdir(parents=True)
            tracked.write_text("fixture")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)
            generated = tracked.with_name("generated.go")
            generated.write_text("generated")
            example = root / "compiler/example/output/generated/client.go"
            example.parent.mkdir(parents=True)
            example.write_text("generated")
            model_build = root / "compiler/example/datamodel/build/client.go"
            model_build.parent.mkdir(parents=True)
            model_build.write_text("generated")
            for name in ("features", ".cursor"):
                notes = root / name / "bin/notes"
                notes.parent.mkdir(parents=True)
                notes.write_text("keep")
            external = Path(tmp) / "outside"
            external.mkdir()
            outside = external / "important"
            outside.write_text("keep")
            (root / "api/build/linked").symlink_to(external, target_is_directory=True)
            cleaner = module.Cleaner(root, dry_run=True)
            cleaner.files()
            self.assertTrue(generated.exists())
            cleaner.dry_run = False
            cleaner.files()
            cleaner.files()
            self.assertFalse(generated.exists())
            self.assertFalse((root / "compiler/example/output").exists())
            self.assertFalse(model_build.parent.exists())
            self.assertEqual(tracked.read_text(), "fixture")
            self.assertEqual(outside.read_text(), "keep")
            for name in ("features", ".cursor"):
                self.assertEqual((root / name / "bin/notes").read_text(), "keep")

    def test_cluster_scope_and_explicit_all(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "nexus-runtime-manifests/kind/.runtime").mkdir(parents=True)
            cleaner = module.Cleaner(root, cluster="selected")
            available = ["runtime", "selected", "nexus-e2e-abcd", "unrelated"]
            self.assertEqual(cleaner.clusters(available), {"runtime", "selected", "nexus-e2e-abcd"})
            cleaner.all_clusters = True
            self.assertEqual(cleaner.clusters(available), set(available))

    def test_runtime_removes_owned_resources_only(self):
        root = Path("/example/repo")
        cleaner = module.Cleaner(root)
        containers = [
            {"Id": "node", "Name": "/e2e-control-plane", "Config": {"Labels": {"io.x-k8s.kind.cluster": "nexus-e2e-abcd"}}},
            {"Id": "foreign", "Name": "/unrelated", "Config": {"Labels": {"io.x-k8s.kind.cluster": "other"}}},
            {"Id": "gw", "Name": "/nexus-api-gw", "Config": {"Labels": {}}},
            {"Id": "etcd", "Name": "/etcd", "Config": {"Labels": {"com.docker.compose.project.working_dir": str(root / "nexus-runtime-manifests/k8s")}}},
        ]
        def query(*args):
            if args == ("docker", "info"): return "ok"
            if args == ("docker", "ps", "-aq"): return "node foreign gw etcd"
            if args[:2] == ("docker", "inspect"): return json.dumps(containers)
            if args[:3] == ("docker", "volume", "ls"): return ""
            if args[:3] == ("docker", "image", "ls"): return "nexus/compiler:latest\nnexus-e2e/gateway:test\nauthz:latest\npostgres:latest\nregistry/nexus/compiler:latest"
            if args[:3] == ("docker", "network", "ls"): return "kind\nnexus"
            if args[:3] == ("docker", "network", "inspect"):
                return json.dumps([{"Containers": {"foreign": {}} if args[-1] == "kind" else {}}])
            self.fail(str(args))
        with patch.object(cleaner, "query", side_effect=query), patch.object(cleaner, "action") as action, patch.object(module.shutil, "which", return_value="tool"):
            cleaner.docker()
        calls = [c.args for c in action.call_args_list]
        self.assertIn(("kind", "delete", "cluster", "--name", "nexus-e2e-abcd"), calls)
        self.assertIn(("docker", "rm", "-fv", "etcd"), calls)
        self.assertIn(("docker", "network", "rm", "nexus"), calls)
        self.assertFalse(any("foreign" in c or "postgres:latest" in c or "registry/nexus/compiler:latest" in c for c in calls))
        self.assertNotIn(("docker", "network", "rm", "kind"), calls)
        self.assertFalse(cleaner.failed)

    def test_failure_is_reported_and_preview_runs_no_mutations(self):
        cleaner = module.Cleaner(Path("/example"), dry_run=True)
        with patch.object(module.subprocess, "run") as run:
            cleaner.action("docker", "rm", "anything")
            run.assert_not_called()
        cleaner.dry_run = False
        with patch.object(module.subprocess, "run", side_effect=subprocess.CalledProcessError(1, "docker")):
            cleaner.action("docker", "rm", "anything")
        self.assertTrue(cleaner.failed)


if __name__ == "__main__":
    unittest.main()
