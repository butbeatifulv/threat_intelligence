#!/usr/bin/env python3
"""Audit agent evaluation registry (GAIA pilot + docs + controls)."""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def main() -> int:
    failed = []
    checks = [
        (ROOT / "eval/gaia/fixtures/pilot/metadata.jsonl", "pilot fixture"),
        (ROOT / "eval/gaia/fixtures/paper-examples/metadata.jsonl", "paper examples"),
        (ROOT / "eval/gaia/paper/system-prompt.txt", "paper system prompt"),
        (ROOT / "eval/gaia/schema/task.schema.json", "task schema"),
        (ROOT / "scripts/eval/gaia/score.py", "scorer"),
        (ROOT / "scripts/eval/gaia/run-pilot.sh", "pilot runner"),
        (ROOT / "docs/agents/agent-evaluation-gaia.md", "GAIA doc"),
        (ROOT / "docs/agents/external-agent-store.md", "agent-store doc"),
        (ROOT / "deploy/eval/gaia.env.example", "env example"),
        (ROOT / ".github/workflows/agent-eval.yml", "agent-eval workflow"),
    ]
    for path, label in checks:
        if not path.exists():
            failed.append(f"missing {label}: {path.relative_to(ROOT)}")

    pilot = ROOT / "eval/gaia/fixtures/pilot/metadata.jsonl"
    if pilot.exists():
        rows = [json.loads(l) for l in pilot.read_text(encoding="utf-8").splitlines() if l.strip()]
        if len(rows) < 3:
            failed.append("pilot fixture needs >= 3 tasks")
        for r in rows:
            for key in ("task_id", "Question", "Level", "Final answer"):
                if key not in r:
                    failed.append(f"pilot row missing {key}: {r.get('task_id')}")

    controls = ROOT / "deploy/security/veil-controls.yaml"
    if controls.exists():
        body = controls.read_text(encoding="utf-8")
        for cid in ("VEIL-EVAL-001", "VEIL-EVAL-002"):
            if cid not in body:
                failed.append(f"veil-controls missing {cid}")
        if "2311.12983" not in body and "arxiv" not in body.lower():
            failed.append("veil-controls should cite arXiv GAIA paper")

    if failed:
        for f in failed:
            print(f"FAIL: {f}")
        return 1
    print(f"agent-eval registry: {len(checks)} artifacts ok")
    return 0


if __name__ == "__main__":
    sys.exit(main())
