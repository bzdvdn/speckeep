#!/usr/bin/env python3
"""Update the Homebrew formula / Scoop manifest for a speckeep release.

Dependency-free so the release runner can call it with plain ``python3``.

Usage::

    update-release.py formula <path> <version> <source_sha256>
    update-release.py scoop   <path> <version> <amd64_sha256> <arm64_sha256>

``version`` is without the leading ``v`` (e.g. ``1.0.1``).
"""

from __future__ import annotations

import json
import re
import sys


def update_formula(path: str, version: str, source_sha: str) -> None:
    text = open(path, encoding="utf-8").read()
    text = re.sub(
        r"(archive/refs/tags/)v[0-9][^\"/]*\.tar\.gz",
        rf"\g<1>v{version}.tar.gz",
        text,
    )
    updated, count = re.subn(
        r'(^\s*sha256\s+")[0-9a-fA-F]{64}(")',
        rf"\g<1>{source_sha}\g<2>",
        text,
        flags=re.MULTILINE,
    )
    if count != 1:
        raise SystemExit(f"expected exactly one sha256 line in {path}, found {count}")
    open(path, "w", encoding="utf-8").write(updated)


def bump_url(url: str, version: str) -> str:
    url = re.sub(r"/v[0-9][^/]*/", f"/v{version}/", url)
    url = re.sub(r"_v[0-9][^_]*(_windows)", rf"_v{version}\g<1>", url)
    return url


def update_scoop(path: str, version: str, amd64: str, arm64: str) -> None:
    with open(path, encoding="utf-8") as handle:
        data = json.load(handle)
    data["version"] = version
    for key, digest in (("64bit", amd64), ("arm64", arm64)):
        entry = data["architecture"][key]
        entry["url"] = bump_url(entry["url"], version)
        entry["hash"] = digest
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(data, handle, indent=2, ensure_ascii=False)
        handle.write("\n")


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        return int(sys.stderr.write(__doc__ or "") or 2)
    kind = argv[1]
    if kind == "formula" and len(argv) == 5:
        update_formula(argv[2], argv[3], argv[4])
        return 0
    if kind == "scoop" and len(argv) == 6:
        update_scoop(argv[2], argv[3], argv[4], argv[5])
        return 0
    sys.stderr.write(__doc__ or "")
    return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
