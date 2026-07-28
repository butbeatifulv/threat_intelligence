#!/usr/bin/env python3
"""Score GAIA-style tasks per arXiv:2311.12983 §3.2 (quasi exact match + normalization)."""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

# Paper: https://arxiv.org/abs/2311.12983 — evaluation via quasi exact match on short factual answers.
FINAL_ANSWER_RE = re.compile(r"final answer:\s*(.+)", re.IGNORECASE | re.DOTALL)


def extract_prediction(text: str) -> str:
    """Strip model reasoning; keep line after FINAL ANSWER: (Figure 2 template)."""
    m = FINAL_ANSWER_RE.search(text or "")
    if m:
        return m.group(1).strip().split("\n")[0].strip()
    return (text or "").strip()


def normalize_answer(text: str) -> str:
    """Light normalization aligned with GAIA leaderboard / paper §3.2."""
    s = extract_prediction(text)
    s = s.strip().lower()
    s = re.sub(r"\s+", " ", s)
    s = re.sub(r"[\"'`]", "", s)
    s = re.sub(r"\.$", "", s)
    # Numbers: drop thousands separators (paper: no commas in numbers unless list)
    if re.fullmatch(r"[+\-]?\d[\d,]*\.?\d*", s.replace(" ", "")):
        s = s.replace(",", "")
    return s


def load_jsonl(path: Path) -> list[dict]:
    rows = []
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line:
            rows.append(json.loads(line))
    return rows


def main() -> int:
    ap = argparse.ArgumentParser(description="Score predictions vs GAIA golden answers")
    ap.add_argument("--tasks", type=Path, required=True, help="metadata.jsonl with Final answer")
    ap.add_argument("--predictions", type=Path, required=True, help="JSONL: task_id, prediction")
    ap.add_argument("--out", type=Path, help="Write metrics JSON")
    args = ap.parse_args()

    tasks = load_jsonl(args.tasks)
    gold = {r["task_id"]: r["Final answer"] for r in tasks}
    level_by_id = {r["task_id"]: int(r.get("Level", 1)) for r in tasks}
    preds: dict[str, str] = {}
    for line in args.predictions.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        row = json.loads(line)
        raw = row.get("prediction", row.get("answer", ""))
        preds[row["task_id"]] = extract_prediction(str(raw))

    correct = 0
    by_level: dict[int, list[bool]] = {}
    details = []
    for tid, ref in gold.items():
        pred = preds.get(tid, "")
        ok = normalize_answer(pred) == normalize_answer(ref)
        if ok:
            correct += 1
        lvl = level_by_id.get(tid, 1)
        by_level.setdefault(lvl, []).append(ok)
        details.append({"task_id": tid, "correct": ok, "level": lvl})

    total = len(gold)
    metrics = {
        "methodology": "arXiv:2311.12983 §3.2 quasi-exact-match",
        "accuracy": correct / total if total else 0.0,
        "correct": correct,
        "total": total,
        "by_level": {
            str(lvl): sum(1 for x in xs if x) / len(xs) if xs else 0.0
            for lvl, xs in sorted(by_level.items())
        },
        "details": details,
    }
    print(json.dumps(metrics, indent=2))
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(json.dumps(metrics, indent=2) + "\n", encoding="utf-8")
    return 0 if correct == total else 1


if __name__ == "__main__":
    sys.exit(main())
