#!/usr/bin/env python3
"""Optional: download GAIA splits from Hugging Face (gated — needs HF_TOKEN).

Normative methodology: https://arxiv.org/abs/2311.12983 — Veil CI does not use this script.
"""
from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
DEFAULT_DIR = ROOT / "eval" / "gaia" / "data"


def main() -> int:
    ap = argparse.ArgumentParser(description="Snapshot gaia-benchmark/GAIA (gated dataset)")
    ap.add_argument("--data-dir", type=Path, default=Path(os.environ.get("GAIA_DATA_DIR", DEFAULT_DIR)))
    ap.add_argument("--config", default=os.environ.get("GAIA_CONFIG", "2023_level1"))
    ap.add_argument("--split", default=os.environ.get("GAIA_SPLIT", "validation"))
    args = ap.parse_args()

    token = os.environ.get("HF_TOKEN") or os.environ.get("HUGGING_FACE_HUB_TOKEN")
    if not token:
        print(
            "HF_TOKEN not set. Veil does not require Hugging Face for agent eval.\n"
            "  Methodology: https://arxiv.org/pdf/2311.12983\n"
            "  Offline CI: make test-agent-eval-pilot test-agent-eval-paper\n"
            "Optional HF download: accept terms at https://huggingface.co/datasets/gaia-benchmark/GAIA\n"
            "  then set HF_TOKEN. Do not commit splits or answers.",
            file=sys.stderr,
        )
        return 2

    try:
        from huggingface_hub import snapshot_download
        from datasets import load_dataset
    except ImportError:
        print("pip install huggingface_hub datasets", file=sys.stderr)
        return 2

    args.data_dir.mkdir(parents=True, exist_ok=True)
    data_dir = snapshot_download(
        repo_id="gaia-benchmark/GAIA",
        repo_type="dataset",
        token=token,
        local_dir=str(args.data_dir / "hf-snapshot"),
    )
    ds = load_dataset(data_dir, args.config, split=args.split)
    out = args.data_dir / f"{args.config}-{args.split}.jsonl"
    with out.open("w", encoding="utf-8") as f:
        for row in ds:
            f.write(
                __import__("json").dumps(
                    {
                        "task_id": row.get("task_id"),
                        "Question": row.get("Question"),
                        "Level": row.get("Level"),
                        "Final answer": row.get("Final answer"),
                        "file_name": row.get("file_name", ""),
                        "file_path": row.get("file_path", ""),
                    },
                    ensure_ascii=False,
                )
                + "\n"
            )
    print(f"wrote {out} ({len(ds)} rows) — keep local; do not commit")
    return 0


if __name__ == "__main__":
    sys.exit(main())
