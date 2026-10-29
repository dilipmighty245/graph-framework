"""Exercise E2E isolation/cleanup without requiring Docker or a cluster."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
FAKE_TOOL = r'''#!/usr/bin/env python3
import json, os, pathlib, sys
name = pathlib.Path(sys.argv[0]).name
args = sys.argv[1:]
state = pathlib.Path(os.environ["HARNESS_STATE"])
with (state / "calls.jsonl").open("a") as f:
    f.write(json.dumps([name] + args) + "\n")
if name == "git":
    if "ls-files" in args:
        sys.stdout.buffer.write(b"Makefile\0")
elif name == "docker":
    if args[0] == "version": print("amd64")
    elif args[0] == "run":
        if os.environ.get("FAIL_STAGE") == "generation": sys.exit(7)
        workdir = args[args.index("--workdir") + 1]
        source = pathlib.Path(next(value.split(":")[0] for index, value in enumerate(args)
                                   if args[index - 1] == "--volume" and value.endswith(":" + workdir)))
        crds = source / "compiler/example/output/generated/crds"
        (source / "compiler/example/tests").mkdir(parents=True)
        crds.mkdir(parents=True)
        (crds / "root.yaml").write_text("kind: CustomResourceDefinition\n")
        runtime = source / "scripts/test-runtime-compatibility.sh"
        runtime.parent.mkdir(parents=True, exist_ok=True)
        runtime.write_text("""#!/usr/bin/env bash
exec python3 - "$0" <<'PY'
import json, os, pathlib, sys
root = pathlib.Path(sys.argv[1]).resolve().parents[1]
assert (root / 'compiler/example/output/generated/crds/root.yaml').exists()
with (pathlib.Path(os.environ['HARNESS_STATE']) / 'calls.jsonl').open('a') as f:
    f.write(json.dumps(['runtime', str(root), os.environ['FRAMEWORK_TEST_RESULTS']]) + '\\n')
sys.exit(11 if os.environ.get('FAIL_STAGE') == 'runtime' else 0)
PY
""")
elif name == "kind":
    if args[0] == "create" and os.environ.get("FAIL_STAGE") == "cluster": sys.exit(8)
elif name == "go":
    if args[0] == "env": print(state / args[1].lower())
    elif os.environ.get("FAIL_STAGE") == "subscription": sys.exit(9)
    elif os.environ.get("FAIL_STAGE") == "streaming" and os.environ.get("KUBE_FEATURE_WatchListClient") == "true": sys.exit(10)
elif name == "kubectl":
    args = args[2:]  # Every operation must use the isolated kubeconfig.
    if args[0] == "create" and args[1] == "-f":
        objects = json.loads(pathlib.Path(args[2]).read_text())["items"]
        (state / "resource.json").write_text(json.dumps(objects[1]))
    elif args[0] == "patch":
        if '"mySlice"' in args[-1]:
            print('spec.testValMarkers.mySlice: Duplicate value: "a"', file=sys.stderr)
            sys.exit(1)
        obj = json.loads((state / "resource.json").read_text())
        obj["spec"]["myStr0"] = "after"
        (state / "resource.json").write_text(json.dumps(obj))
    elif args[0] == "delete": (state / "deleted").touch()
    elif args[0] == "get" and args[1].startswith("configs.") and not (state / "deleted").exists():
        obj = json.loads((state / "resource.json").read_text())
        if os.environ.get("FAIL_STAGE") == "read": obj["spec"]["myStr0"] = "wrong"
        print(json.dumps(obj))
'''


class RunnerLifecycleTests(unittest.TestCase):
    def run_runner(self, fail_stage="", compare_streaming=False, runtime=False):
        temp = tempfile.TemporaryDirectory(prefix="framework-harness-")
        self.addCleanup(temp.cleanup)
        state = Path(temp.name)
        tools = state / "bin"
        tools.mkdir()
        for name in ("git", "docker", "kind", "kubectl", "go"):
            tool = tools / name
            tool.write_text(FAKE_TOOL)
            tool.chmod(0o755)
        kubeconfig = state / "user-kubeconfig"
        kubeconfig.write_text("unchanged")
        env = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ["PATH"],
                   HARNESS_STATE=str(state), TMPDIR=str(state),
                   FRAMEWORK_TEST_RESULTS=str(state / "results"),
                   E2E_COMPONENTS="resources", E2E_COMPILER_IMAGE="test/compiler:local",
                   E2E_SDK_RUNS="1",
                   E2E_SUBSCRIPTION_COMPARE_STREAMING=str(compare_streaming).lower(),
                   E2E_RUNTIME_COMPATIBILITY=str(runtime).lower(),
                   KUBECONFIG=str(kubeconfig), FAIL_STAGE=fail_stage)
        if compare_streaming is None:
            env.pop("E2E_SUBSCRIPTION_COMPARE_STREAMING")
        result = subprocess.run(["bash", str(ROOT / "scripts/test-framework-e2e.sh")],
                                env=env, capture_output=True, text=True, timeout=20)
        calls = [json.loads(line) for line in (state / "calls.jsonl").read_text().splitlines()]
        self.assertEqual(kubeconfig.read_text(), "unchanged")
        self.assertEqual(list(state.glob("nexus-e2e-*")), [], "temporary source should be removed")
        for call in calls:
            if call[0] == "kubectl":
                self.assertEqual(call[1], "--kubeconfig")
                self.assertNotEqual(call[2], str(kubeconfig))
        return result, calls

    def test_resource_lifecycle_and_cleanup(self):
        result, calls = self.run_runner()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertTrue(any(c[:2] == ["kind", "create"] for c in calls))
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))
        self.assertTrue(any(c[0] == "kubectl" and c[3] == "patch" for c in calls))
        self.assertTrue(any(c[0] == "kubectl" and c[3] == "delete" for c in calls))

    def test_generation_failure_never_creates_cluster(self):
        result, calls = self.run_runner("generation")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(any(c[0] == "kind" for c in calls))

    def test_partial_cluster_creation_is_cleaned_up(self):
        result, calls = self.run_runner("cluster")
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))

    def test_resource_assertion_failure_is_not_hidden(self):
        result, calls = self.run_runner("read")
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))

    def test_subscription_failure_is_reported_and_cluster_is_removed(self):
        result, calls = self.run_runner("subscription")
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(any(c[0] == "go" and "-tags=integration" in c for c in calls))
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))

    def test_streaming_comparison_uses_both_modes_without_module_edits(self):
        result, calls = self.run_runner(compare_streaming=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(sum(c[0] == "go" and "-tags=integration" in c for c in calls), 2)
        self.assertEqual(sum(c[:3] == ["go", "mod", "edit"] for c in calls), 0)
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))

    def test_default_runs_both_modes_with_readonly_dependencies(self):
        result, calls = self.run_runner(compare_streaming=None)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        subscriptions = [c for c in calls if c[0] == "go" and "-tags=integration" in c]
        self.assertEqual(len(subscriptions), 2)
        self.assertTrue(all("-mod=readonly" in c for c in subscriptions))

    def test_runtime_uses_fresh_snapshot_and_persistent_results(self):
        result, calls = self.run_runner(runtime=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        runtime = [c for c in calls if c[0] == "runtime"]
        self.assertEqual(len(runtime), 1)
        self.assertNotEqual(runtime[0][1], str(ROOT))
        self.assertTrue(Path(runtime[0][2]).name.startswith("e2e-"))

    def test_runtime_failure_is_reported_and_cluster_is_removed(self):
        result, calls = self.run_runner("runtime", runtime=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(any(c[0] == "runtime" for c in calls))
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))

    def test_streaming_failure_is_reported_and_cluster_is_removed(self):
        result, calls = self.run_runner("streaming", compare_streaming=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(sum(c[0] == "go" and "-tags=integration" in c for c in calls), 2)
        self.assertTrue(any(c[:2] == ["kind", "delete"] for c in calls))


if __name__ == "__main__":
    unittest.main()
