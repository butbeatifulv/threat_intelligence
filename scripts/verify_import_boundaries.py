#!/usr/bin/env python3
"""Verify Veil Go layer import boundaries (discovery / pipeline / knowledge isolation)."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LAYERS = ("discovery", "pipeline", "knowledge")
FORBIDDEN_CROSS = {
    "discovery": ("pipeline/", "knowledge/"),
    "pipeline": ("discovery/", "knowledge/"),
    "knowledge": ("discovery/", "pipeline/"),
}
IO_IMPORTS = re.compile(
    r'"github\.com/(neo4j|nats-io)|"net/http"|"github\.com/butbeautifulv/veil/knowledge/connector/neo4j"'
)


def layer_of(path: Path) -> str | None:
    rel = path.relative_to(ROOT).as_posix()
    for layer in LAYERS:
        if rel.startswith(layer + "/"):
            return layer
    return None


def check_cross_imports() -> list[str]:
    errors: list[str] = []
    for go_file in ROOT.rglob("*.go"):
        if "/vendor/" in go_file.as_posix():
            continue
        layer = layer_of(go_file)
        if layer is None:
            continue
        text = go_file.read_text(encoding="utf-8", errors="replace")
        for imp in FORBIDDEN_CROSS[layer]:
            if f"github.com/butbeautifulv/veil/{imp}" in text:
                errors.append(f"{go_file}: cross-layer import forbidden ({layer} -> {imp})")
    return errors


def check_domain_io() -> list[str]:
    errors: list[str] = []
    for go_file in ROOT.rglob("*.go"):
        if "/domain/" not in go_file.as_posix():
            continue
        text = go_file.read_text(encoding="utf-8", errors="replace")
        if IO_IMPORTS.search(text):
            errors.append(f"{go_file}: domain package must not import I/O drivers")
    return errors


def main() -> int:
    errors = check_cross_imports() + check_domain_io()
    if errors:
        print("import boundary violations:", file=sys.stderr)
        for e in errors:
            print(f"  - {e}", file=sys.stderr)
        return 1
    print("import boundaries OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
