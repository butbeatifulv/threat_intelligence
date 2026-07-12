#!/usr/bin/env python3
"""Export TI/vuln text chunks from Neo4j for hybrid search indexing."""
from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path


def repo_root() -> Path:
    here = Path(__file__).resolve()
    for parent in [here.parent, *here.parents]:
        if (parent / "versions.env").is_file():
            return parent
    return Path.cwd()


def export_via_bolt(uri: str, user: str, password: str, database: str) -> tuple[list[dict], list[dict]]:
    try:
        from neo4j import GraphDatabase
    except ImportError:
        print("ERROR: pip install neo4j", file=sys.stderr)
        sys.exit(1)
    ti_chunks: list[dict] = []
    vuln_chunks: list[dict] = []
    driver = GraphDatabase.driver(uri, auth=(user, password))
    with driver.session(database=database) as session:
        for row in session.run(
            """
            MATCH (n)
            WHERE any(l IN labels(n) WHERE l IN ['ThreatActor','Campaign','Malware','Tool','IntrusionSet'])
            AND (n.description IS NOT NULL OR n.name IS NOT NULL OR n.title IS NOT NULL)
            RETURN coalesce(n.id, n.name, n.title) AS id,
                   coalesce(n.title, n.name) AS title,
                   coalesce(n.description, '') AS description,
                   labels(n) AS labels
            LIMIT 50000
            """
        ):
            text = f"{row['title'] or ''} {row['description'] or ''}".strip()
            if not text:
                continue
            ti_chunks.append(
                {
                    "chunk_id": f"ti::{row['id']}",
                    "skill_id": str(row["id"]),
                    "node_id": str(row["id"]),
                    "category": "ti",
                    "kind": (row["labels"] or ["TI"])[0],
                    "text": text[:4000],
                }
            )
        for row in session.run(
            """
            MATCH (n:Vulnerability)
            WHERE n.cve IS NOT NULL OR n.description IS NOT NULL
            RETURN coalesce(n.cve, n.id) AS id,
                   coalesce(n.cve, n.title, n.name) AS title,
                   coalesce(n.description, '') AS description
            LIMIT 50000
            """
        ):
            text = f"{row['title'] or ''} {row['description'] or ''}".strip()
            if not text:
                continue
            vuln_chunks.append(
                {
                    "chunk_id": f"vuln::{row['id']}",
                    "skill_id": str(row["id"]),
                    "node_id": str(row["id"]),
                    "category": "vuln",
                    "kind": "Vulnerability",
                    "text": text[:4000],
                }
            )
    driver.close()
    return ti_chunks, vuln_chunks


def write_jsonl(path: Path, rows: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out-dir", default="docs/skills-index")
    ap.add_argument("--uri", default=os.environ.get("NEO4J_URI", "neo4j://127.0.0.1:7687"))
    ap.add_argument("--user", default=os.environ.get("NEO4J_USER", "neo4j"))
    ap.add_argument("--password", default=os.environ.get("NEO4J_PASS", "neo4jpassword"))
    ap.add_argument("--database", default=os.environ.get("NEO4J_DB", "neo4j"))
    args = ap.parse_args()

    root = repo_root()
    out = root / args.out_dir
    ti, vuln = export_via_bolt(args.uri, args.user, args.password, args.database)
    write_jsonl(out / "ti-chunks.jsonl", ti)
    write_jsonl(out / "vuln-chunks.jsonl", vuln)
    print(f"OK: ti={len(ti)} vuln={len(vuln)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
