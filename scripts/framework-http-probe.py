#!/usr/bin/env python3
"""Wait for actual gateway/GraphQL data, failing on HTTP or application errors."""
import argparse
import json
import time
import urllib.error
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument("mode", choices=["gateway", "graphql"])
parser.add_argument("url")
parser.add_argument("--name", required=True)
parser.add_argument("--value", default="after")
parser.add_argument("--timeout", type=float, default=90)
args = parser.parse_args()
deadline = time.monotonic() + args.timeout
last_error = "no response"
while time.monotonic() < deadline:
    try:
        if args.mode == "graphql":
            body = json.dumps({"query": "{ root { Id Config { MyStr0 } } }"}).encode()
            request = urllib.request.Request(args.url, body, {"Content-Type": "application/json"})
        else:
            request = urllib.request.Request(args.url)
        with urllib.request.urlopen(request, timeout=3) as response:
            data = json.load(response)
        if args.mode == "graphql":
            assert not data.get("errors"), data.get("errors")
            root = data["data"]["root"]
            assert root["Id"] == "default", root
            # Cross-package aliases use JSON scalar strings in this schema.
            assert json.loads(root["Config"]["MyStr0"]) == args.value, root
        else:
            assert data["metadata"]["name"] == args.name, data
            assert data["spec"]["myStr0"] == args.value, data
        print(f"PASS: {args.mode} returned the generated resource and expected value")
        break
    except (OSError, ValueError, KeyError, TypeError, AssertionError) as exc:
        last_error = str(exc)
        time.sleep(1)
else:
    raise SystemExit(f"FAIL: {args.mode} probe timed out: {last_error}")
