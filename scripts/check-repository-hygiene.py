#!/usr/bin/env python3
"""Reject retired corporate source and registry endpoints in project files."""
import os
from pathlib import Path
import re
import subprocess

root = Path(__file__).resolve().parents[1]
patterns = [
    r"intel" + r"-innersource",
    r"intel" + r"-sandbox",
    r"proxy-dmz" + r"\.intel\.com",
    r"github\.com/" + r"vmware-tanzu/",
    r"gitlab\.(?:eng\.)?vmware\.com",
    r"amr-registry[^/\s]*\.intel\.com",
    r"caas\.intel\.com",
]
retired = re.compile("|".join(patterns), re.IGNORECASE)
files = subprocess.check_output(
    ["git", "-C", str(root), "ls-files", "-z", "--cached", "--others", "--exclude-standard"]
)
failures = []
for raw in set(files.split(b"\0")):
    if not raw:
        continue
    relative = Path(os.fsdecode(raw))
    if relative.parts[0] in (".cursor", "features"):
        continue
    path = root / relative
    if not path.is_file():
        continue
    content = path.read_bytes()
    if b"\0" in content:
        continue
    for number, line in enumerate(content.decode("utf-8", errors="replace").splitlines(), 1):
        if retired.search(line):
            failures.append(f"{relative}:{number}: retired corporate endpoint")
if failures:
    raise SystemExit("\n".join(sorted(failures)))
print("PASS: no retired corporate source or registry endpoints")
