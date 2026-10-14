"""Local builds must not assume that project images have been published."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class ImageConfigurationTests(unittest.TestCase):
    def image(self, directory, variable, *settings):
        result = subprocess.run(
            ["make", "--no-print-directory", "-f", "Makefile", "-f", "-", "print-test-image", *settings],
            cwd=ROOT / directory,
            input=f'print-test-image:\n\t@printf "%s\\n" "$({variable})"\n',
            text=True, capture_output=True, timeout=15,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout.strip()

    def test_root_builds_use_local_tags_unless_registry_is_supplied(self):
        for registry, expected in [("", "nexus/api-gw:check"),
                                   ("registry.invalid/team", "registry.invalid/team/nexus/api-gw:check"),
                                   ("registry.invalid/team/", "registry.invalid/team/nexus/api-gw:check")]:
            with self.subTest(registry=registry):
                self.assertEqual(self.image(".", "API_GW_DOCKER_IMAGE", "TAG=check", f"DOCKER_REGISTRY={registry}"), expected)

    def test_authorization_default_is_a_local_image(self):
        self.assertEqual(self.image("authz-controller", "DOCKER_IMAGE_PREFIX", "DOCKER_REGISTRY="), "")
        result = subprocess.run(["make", "--no-print-directory", "show-image-name", "DOCKER_REGISTRY=", "TAG=check"],
                                cwd=ROOT / "authz-controller", capture_output=True, text=True, timeout=15)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), "authz:check")

    def test_datamodel_compiler_tags_have_no_leading_slash(self):
        for directory, registry in [("api", "COMPILER_DOCKER_REGISTRY="),
                                    ("compiler/example/datamodel", "COMPILER_DOCKER_REGISTRY="),
                                    ("demo/traffic-lights/datamodel", "COMPILER_DOCKER_REPO=")]:
            with self.subTest(directory=directory):
                self.assertEqual(self.image(directory, "COMPILER_DOCKER_TAG", registry, "COMPILER_TAG=check", "GOARCH=amd64"),
                                 "nexus/compiler/amd64:check")

    def test_download_without_registry_fails_before_docker_is_called(self):
        with tempfile.TemporaryDirectory(prefix="registry-check-") as temp:
            tools = Path(temp)
            marker = tools / "docker-called"
            docker = tools / "docker"
            docker.write_text('#!/bin/sh\ntouch "$DOCKER_CALLED"\n')
            docker.chmod(0o755)
            env = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ["PATH"], DOCKER_CALLED=str(marker))
            result = subprocess.run(["make", "--no-print-directory", "core.runtime.download", "DOCKER_REGISTRY=", "TAG=check"],
                                    cwd=ROOT, env=env, capture_output=True, text=True, timeout=15)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("Set DOCKER_REGISTRY explicitly", result.stderr)
            self.assertFalse(marker.exists())


if __name__ == "__main__":
    unittest.main()
