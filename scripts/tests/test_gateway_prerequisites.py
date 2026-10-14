"""Missing generated APIs must fail before Go dependency resolution."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class GatewayPrerequisites(unittest.TestCase):
    def test_missing_api_reports_generation_step_without_running_go(self):
        with tempfile.TemporaryDirectory(prefix="gateway-prereq-") as temp:
            root = Path(temp)
            (root / "scripts").mkdir()
            shutil.copy2(ROOT / "scripts/check-gateway.sh", root / "scripts/check-gateway.sh")
            tools = root / "bin"
            tools.mkdir()
            marker = root / "go-called"
            fake_go = tools / "go"
            fake_go.write_text('#!/bin/sh\ntouch "$GO_CALLED"\nexit 99\n')
            fake_go.chmod(0o755)
            env = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ["PATH"], GO_CALLED=str(marker))
            result = subprocess.run(["bash", str(root / "scripts/check-gateway.sh")], env=env,
                                    capture_output=True, text=True, timeout=10)
            self.assertEqual(result.returncode, 2)
            self.assertIn("Gateway prerequisite missing", result.stderr)
            self.assertIn("make -C api datamodel_build", result.stderr)
            self.assertFalse(marker.exists())


if __name__ == "__main__":
    unittest.main()
