#!/usr/bin/env python3
"""Build docs/skills-index/cyber-skills.json from .external/Anthropic-Cybersecurity-Skills-main."""
from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import datetime, timezone
from pathlib import Path

ATTACK_RE = re.compile(r"\bT\d{4}(?:\.\d{3})?\b", re.I)
FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)


def repo_root() -> Path:
    here = Path(__file__).resolve()
    for parent in [here.parent, *here.parents]:
        if (parent / "versions.env").is_file() and (parent / "go.mod").exists():
            return parent
    return Path.cwd()


def parse_simple_yaml(block: str) -> dict:
    data: dict = {}
    key: str | None = None
    list_key: str | None = None
    for line in block.splitlines():
        if not line.strip() or line.strip().startswith("#"):
            continue
        if line.startswith("  - ") and list_key:
            data.setdefault(list_key, []).append(line[4:].strip().strip("'\""))
            continue
        m = re.match(r"^([a-zA-Z0-9_]+):\s*(.*)$", line)
        if not m:
            continue
        key, val = m.group(1), m.group(2).strip()
        list_key = None
        if val == "":
            list_key = key
            data[key] = []
            continue
        if val.startswith("[") and val.endswith("]"):
            inner = val[1:-1].strip()
            data[key] = [x.strip().strip("'\"") for x in inner.split(",") if x.strip()] if inner else []
        elif val.startswith("'") or val.startswith('"'):
            data[key] = val.strip("'\"")
        else:
            data[key] = val
    return data


def extract_attack_ids(text: str) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for m in ATTACK_RE.finditer(text):
        tid = m.group(0).upper()
        if tid not in seen:
            seen.add(tid)
            out.append(tid)
    return sorted(out)


def skill_from_md(repo: Path, skill_dir: Path) -> dict | None:
    md = skill_dir / "SKILL.md"
    if not md.is_file():
        return None
    raw = md.read_text(encoding="utf-8", errors="replace")
    meta: dict = {}
    body = raw
    fm = FRONTMATTER_RE.match(raw)
    if fm:
        meta = parse_simple_yaml(fm.group(1))
        body = raw[fm.end() :]
    slug = skill_dir.name
    name = str(meta.get("name") or slug)
    attack_ids = extract_attack_ids(raw)
    rel = md.relative_to(repo).as_posix()
    return {
        "id": slug,
        "name": name,
        "domain": str(meta.get("domain") or "cybersecurity"),
        "subdomain": str(meta.get("subdomain") or ""),
        "description": str(meta.get("description") or "")[:500],
        "tags": meta.get("tags") if isinstance(meta.get("tags"), list) else [],
        "nist_csf": meta.get("nist_csf") if isinstance(meta.get("nist_csf"), list) else [],
        "version": str(meta.get("version") or ""),
        "license": str(meta.get("license") or "Apache-2.0"),
        "attack_ids": attack_ids,
        "corpus_path": rel,
        "external_path": rel,
        "body_chars": len(body),
    }


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "--corpus-skills",
        default="corpus/anthropic-cybersecurity-skills/skills",
        help="Committed skills tree (relative to repo root)",
    )
    ap.add_argument(
        "--external",
        default=".external/Anthropic-Cybersecurity-Skills-main",
        help="Fallback dev import path when --corpus-skills missing",
    )
    ap.add_argument(
        "--out",
        default="docs/skills-index/cyber-skills.json",
        help="Output JSON path (relative to repo root)",
    )
    ap.add_argument("--emit-cypher", metavar="FILE", help="Write Neo4j seed Cypher to FILE")
    ap.add_argument(
        "--check",
        action="store_true",
        help="Exit 1 if output JSON would change (CI stale check)",
    )
    args = ap.parse_args()

    root = repo_root()
    skills_root = (root / args.corpus_skills).resolve()
    source_path = args.corpus_skills
    if not skills_root.is_dir():
        external = (root / args.external).resolve()
        skills_root = external / "skills"
        source_path = args.external
    if not skills_root.is_dir():
        if args.check:
            out_path = root / args.out
            if not out_path.is_file():
                print(f"ERROR: missing index {out_path}", file=sys.stderr)
                return 1
            try:
                doc = json.loads(out_path.read_text(encoding="utf-8"))
            except json.JSONDecodeError as e:
                print(f"ERROR: invalid index JSON: {e}", file=sys.stderr)
                return 1
            n = int(doc.get("skill_count") or len(doc.get("skills") or []))
            if n < 700:
                print(f"ERROR: index has only {n} skills", file=sys.stderr)
                return 1
            print(f"OK: committed index ({n} skills); .external missing — full regen skipped")
            return 0
        print(f"ERROR: skills dir not found: {skills_root}", file=sys.stderr)
        print("Run: make corpus-import  (or clone under .external/)", file=sys.stderr)
        return 1

    entries: list[dict] = []
    for skill_dir in sorted(skills_root.iterdir()):
        if not skill_dir.is_dir():
            continue
        row = skill_from_md(root, skill_dir)
        if row:
            entries.append(row)

    all_attack: set[str] = set()
    subdomains: dict[str, int] = {}
    for e in entries:
        for tid in e["attack_ids"]:
            all_attack.add(tid)
        sub = e.get("subdomain") or "unknown"
        subdomains[sub] = subdomains.get(sub, 0) + 1

    doc = {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%MZ"),
        "source": "Anthropic-Cybersecurity-Skills",
        "source_path": source_path,
        "corpus_skills_path": args.corpus_skills,
        "mappings_path": "pkg/playbook/corpus/mappings",
        "skill_count": len(entries),
        "unique_attack_ids": len(all_attack),
        "subdomain_counts": dict(sorted(subdomains.items(), key=lambda x: (-x[1], x[0]))),
        "skills": entries,
    }

    out_path = root / args.out
    new_body = json.dumps(doc, indent=2, ensure_ascii=False) + "\n"
    if args.check:
        if not out_path.is_file():
            print(f"ERROR: missing index {out_path}; run make skills-index", file=sys.stderr)
            return 1
        try:
            existing_doc = json.loads(out_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as e:
            print(f"ERROR: invalid index JSON: {e}", file=sys.stderr)
            return 1
        doc.pop("generated_at", None)
        existing_doc.pop("generated_at", None)
        if existing_doc != doc:
            print(f"ERROR: stale index {out_path}; run make skills-index", file=sys.stderr)
            return 1
        print(f"OK: {len(entries)} skills index up to date")
        return 0
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(new_body, encoding="utf-8")
    print(f"Wrote {len(entries)} skills -> {out_path}")

    if args.emit_cypher:
        cypher_path = root / args.emit_cypher
        cypher_path.parent.mkdir(parents=True, exist_ok=True)
        lines = [
            "// Generated by scripts/knowledge/generate-cyber-skills-index.py --emit-cypher",
            "// MERGE CyberSkill nodes and HAS_PLAYBOOK edges to AttackTechnique",
            "",
        ]
        for e in entries:
            sid = e["id"].replace("\\", "\\\\").replace("'", "\\'")
            title = e["name"].replace("'", "\\'")[:200]
            sub = (e.get("subdomain") or "").replace("'", "\\'")
            lines.append(
                f"MERGE (s:CyberSkill {{id: '{sid}'}}) "
                f"SET s.title = '{title}', s.subdomain = '{sub}', s.source = 'anthropic-cyber-skills';"
            )
            for tid in e["attack_ids"]:
                lines.append(
                    f"MATCH (t:AttackTechnique {{id: '{tid}'}}), (s:CyberSkill {{id: '{sid}'}}) "
                    f"MERGE (t)-[:HAS_PLAYBOOK]->(s);"
                )
        cypher_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
        print(f"Wrote Cypher -> {cypher_path}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
