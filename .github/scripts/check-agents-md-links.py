#!/usr/bin/env python3
"""Verify that relative Markdown links in AGENTS.md resolve to real paths.

AGENTS.md is the primary agent navigation file. Broken relative links send an
agent to dead ends, so this check fails CI when any relative link target is
missing. External URLs (http/https/mailto) and pure anchors are ignored; only
repository-relative file or directory links are validated.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TARGET = ROOT / "AGENTS.md"

# Matches Markdown links and images: [text](target) / ![alt](target)
LINK_RE = re.compile(r"!?\[[^\]]*\]\(([^)]+)\)")


def is_external(target: str) -> bool:
    return target.startswith(
        ("http://", "https://", "mailto:", "tel:", "#", "//")
    )


def link_targets(text: str):
    for match in LINK_RE.finditer(text):
        raw = match.group(1).strip()
        if not raw:
            continue
        # Drop an optional "title": [t](path "title")
        token = raw.split()[0]
        # Strip angle brackets: [t](<path>)
        token = token.strip("<>")
        # Drop anchor / query fragments.
        token = token.split("#", 1)[0].split("?", 1)[0]
        if token:
            yield token


def main() -> int:
    if not TARGET.exists():
        print(f"error: {TARGET} not found", file=sys.stderr)
        return 2
    text = TARGET.read_text(encoding="utf-8")
    missing = []
    checked = 0
    for target in link_targets(text):
        if is_external(target):
            continue
        checked += 1
        if not (ROOT / target).exists():
            missing.append(target)
    if missing:
        print("AGENTS.md has broken relative links:")
        for link in sorted(set(missing)):
            print(f"  - {link}")
        return 1
    print(f"AGENTS.md: all {checked} relative links resolve.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
