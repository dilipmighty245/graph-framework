"""Ensure HTTP success alone cannot make a service probe pass."""
import contextlib
import io
from pathlib import Path
import runpy
import unittest
from unittest.mock import patch

PROBE = Path(__file__).resolve().parents[1] / "framework-http-probe.py"


class ResponseChecks(unittest.TestCase):
    def probe(self, mode, response):
        import json
        with patch("sys.argv", [str(PROBE), mode, "http://localhost/", "--name", "test", "--timeout", "1"]), \
             patch("urllib.request.urlopen") as urlopen, \
             patch("time.monotonic", side_effect=[0, 0, 2]), \
             patch("time.sleep"), contextlib.redirect_stdout(io.StringIO()):
            urlopen.return_value.__enter__.return_value = io.StringIO(json.dumps(response))
            runpy.run_path(str(PROBE), run_name="__main__")
            return urlopen.call_args.args[0]

    def test_gateway_checks_resource_identity_and_value(self):
        self.probe("gateway", {"metadata": {"name": "test"}, "spec": {"myStr0": "after"}})
        with self.assertRaises(SystemExit):
            self.probe("gateway", {"metadata": {"name": "other"}, "spec": {"myStr0": "after"}})

    def test_graphql_checks_backend_resource_data(self):
        request = self.probe("graphql", {"data": {"root": {"Id": "default", "Config": {"MyStr0": '"after"'}}}})
        self.assertEqual(request.get_method(), "POST")
        with self.assertRaises(SystemExit):
            self.probe("graphql", {"data": {"root": {"Id": "default", "Config": {"MyStr0": '"before"'}}}})

    def test_graphql_http_success_with_errors_fails(self):
        with self.assertRaises(SystemExit):
            self.probe("graphql", {"errors": [{"message": "resolver failed"}], "data": None})


if __name__ == "__main__":
    unittest.main()
