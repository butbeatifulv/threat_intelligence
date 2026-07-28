#!/usr/bin/env python3
"""Build docs/skills-index/procedures-index.json from committed corpus SKILL.md files."""
from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import datetime, timezone
from pathlib import Path

# Shared with generate-cyber-skills-index.py
FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)
SECTION_RE = re.compile(r"^##\s+(.+)$", re.MULTILINE)
STEP_RE = re.compile(r"^###\s+Step\s+(\d+):\s*(.+)$", re.MULTILINE)
ATTACK_RE = re.compile(r"\bT\d{4}(?:\.\d{3})?\b", re.I)
# Common CLI / product tokens for catalog linking
TOOL_TOKEN_RE = re.compile(
    r"\b(nmap|nuclei|masscan|nikto|sqlmap|gobuster|ffuf|feroxbuster|"
    r"burp|wireshark|volatility|yara|sigma|splunk|elastic|misp|"
    r"theharvester|amass|subfinder|httpx|wpscan|metasploit|bloodhound|"
    r"hashcat|john|hydra|responder|impacket|shodan|censys|trivy|grype|"
    r"kubesec|falco|osquery|velociraptor|autopsy|dcfldd|dd)\b",
    re.I,
)


def repo_root() -> Path:
    here = Path(__file__).resolve()
    for parent in [here.parent, *here.parents]:
        if (parent / "versions.env").is_file() and (parent / "go.mod").exists():
            return parent
    return Path.cwd()


def parse_simple_yaml(block: str) -> dict:
    data: dict = {}
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


def load_catalog_names(root: Path) -> set[str]:
    """Tool-level catalog names only (exclude YAML parameter `name:` fields)."""
    names: set[str] = set()
    for path in [
        root / "engage/serve/catalog/tools.yaml",
        root / "engage/serve/catalog/tools.live.yaml",
    ]:
        if not path.is_file():
            continue
        for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
            m = re.match(r"^  - name:\s+(\S+)", line)
            if m:
                names.add(m.group(1))
            m2 = re.match(r"^    binary:\s+(\S+)", line)
            if m2:
                names.add(m2.group(1))
    return names


def map_token_to_catalog(token: str, catalog: set[str]) -> str | None:
    t = token.lower()
    aliases = {
        "nmap": "nmap_scan",
        "nuclei": "nuclei_scan",
        "nikto": "nikto_scan",
        "sqlmap": "sqlmap_scan",
        "gobuster": "gobuster_scan",
        "ffuf": "ffuf_scan",
        "feroxbuster": "feroxbuster_scan",
        "wpscan": "wpscan_scan",
        "masscan": "masscan_scan",
        "httpx": "httpx_probe",
        "subfinder": "subfinder_scan",
        "amass": "amass_enum",
        "theharvester": "theharvester_osint",
    }
    if t in aliases and aliases[t] in catalog:
        return aliases[t]
    if len(t) < 3:
        return None
    for name in sorted(catalog):
        low = name.lower()
        if low == t or low.endswith(f"_{t}") or low.startswith(f"{t}_"):
            return name
    candidate = f"{t}_scan"
    if candidate in catalog:
        return candidate
    return None


def extract_bullets(block: str) -> list[str]:
    out: list[str] = []
    for line in block.splitlines():
        line = line.strip()
        if line.startswith("- "):
            out.append(line[2:].strip())
    return out


def split_sections(body: str) -> dict[str, str]:
    sections: dict[str, str] = {}
    matches = list(SECTION_RE.finditer(body))
    for i, m in enumerate(matches):
        title = m.group(1).strip().lower()
        start = m.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(body)
        sections[title] = body[start:end].strip()
    return sections


def classify_step(body: str) -> str:
    if "```bash" in body or "```sh" in body or body.strip().startswith("#"):
        return "shell"
    return "manual"


def extract_steps(workflow: str) -> list[dict]:
    steps: list[dict] = []
    parts = STEP_RE.split(workflow)
    if len(parts) <= 1:
        return steps
    i = 1
    while i + 2 <= len(parts):
        num = int(parts[i])
        title = parts[i + 1].strip()
        content = parts[i + 2].strip() if i + 2 < len(parts) else ""
        mentions = sorted({m.lower() for m in TOOL_TOKEN_RE.findall(content)})
        steps.append({
            "number": num,
            "title": title,
            "kind": classify_step(content),
            "body": content[:2000],
            "tool_mentions": mentions,
        })
        i += 3
    return steps


def procedure_from_skill(repo: Path, skill_dir: Path, catalog: set[str]) -> dict | None:
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
    sections = split_sections(body)
    when = extract_bullets(sections.get("when to use", ""))
    prereq = extract_bullets(sections.get("prerequisites", ""))
    workflow = sections.get("workflow", "") or sections.get("detection workflow", "")
    steps = extract_steps(workflow)
    scenarios = extract_bullets(sections.get("scenarios", ""))
    all_mentions: set[str] = set()
    for s in steps:
        all_mentions.update(s.get("tool_mentions", []))
    tools_block = sections.get("tools & systems", "") or sections.get("tools", "")
    all_mentions.update(m.lower() for m in TOOL_TOKEN_RE.findall(tools_block))
    catalog_tools: list[str] = []
    seen_cat: set[str] = set()
    for m in sorted(all_mentions):
        hit = map_token_to_catalog(m, catalog)
        if hit and hit not in seen_cat:
            seen_cat.add(hit)
            catalog_tools.append(hit)
    attack_ids = sorted({m.group(0).upper() for m in ATTACK_RE.finditer(raw)})
    rel = md.relative_to(repo).as_posix()
    return {
        "id": slug,
        "subdomain": str(meta.get("subdomain") or ""),
        "attack_ids": attack_ids,
        "nist_csf": meta.get("nist_csf") if isinstance(meta.get("nist_csf"), list) else [],
        "step_count": len(steps),
        "when_to_use_count": len(when),
        "prereq_count": len(prereq),
        "tool_mentions": sorted(all_mentions),
        "catalog_tools": catalog_tools,
        "corpus_path": rel,
        "_when_to_use": when,
        "_prerequisites": prereq,
        "_steps": steps,
        "_scenarios": scenarios,
    }


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "--corpus-skills",
        default="corpus/anthropic-cybersecurity-skills/skills",
    )
    ap.add_argument("--out", default="docs/skills-index/procedures-index.json")
    ap.add_argument("--matrix-out", default="docs/playbooks/playbook-import-matrix.md")
    ap.add_argument("--check", action="store_true")
    args = ap.parse_args()

    root = repo_root()
    skills_root = (root / args.corpus_skills).resolve()
    if not skills_root.is_dir():
        print(f"ERROR: {skills_root} not found; run make corpus-import", file=sys.stderr)
        return 1

    catalog = load_catalog_names(root)
    entries: list[dict] = []
    by_sub: dict[str, list[dict]] = {}
    for skill_dir in sorted(skills_root.iterdir()):
        if not skill_dir.is_dir():
            continue
        row = procedure_from_skill(root, skill_dir, catalog)
        if not row:
            continue
        pub = {k: v for k, v in row.items() if not k.startswith("_")}
        entries.append(pub)
        sub = pub.get("subdomain") or "unknown"
        by_sub.setdefault(sub, []).append(pub)

    doc = {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%MZ"),
        "skill_count": len(entries),
        "procedures": entries,
    }
    out_path = root / args.out
    new_body = json.dumps(doc, indent=2, ensure_ascii=False) + "\n"

    if args.check:
        if not out_path.is_file():
            print(f"ERROR: missing {out_path}", file=sys.stderr)
            return 1
        existing = json.loads(out_path.read_text(encoding="utf-8"))
        existing.pop("generated_at", None)
        check_doc = dict(doc)
        check_doc.pop("generated_at", None)
        if existing != check_doc:
            print(f"ERROR: stale {out_path}; run make procedures-index", file=sys.stderr)
            return 1
        print(f"OK: {len(entries)} procedures index up to date")
        return 0

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(new_body, encoding="utf-8")
    print(f"Wrote {len(entries)} procedures -> {out_path}")

    # Import matrix markdown
    priority = {
        "digital-forensics": "P1",
        "incident-response": "P1",
        "threat-hunting": "P1",
        "malware-analysis": "P2",
        "penetration-testing": "P2",
        "web-application-security": "P2",
        "vulnerability-management": "P2",
        "threat-intelligence": "P2",
    }
    lines = [
        "# Playbook import matrix",
        "",
        "Tracker for migrating Anthropic skills into the Veil Knowledge domain. "
        "Regenerate with `make procedures-index`.",
        "",
        "| Priority | Subdomain | Skills | Procedures | Catalog linked | Status | Batch |",
        "|----------|-----------|--------|------------|----------------|--------|-------|",
    ]
    batch = 1
    for sub in sorted(by_sub.keys(), key=lambda s: (-len(by_sub[s]), s)):
        rows = by_sub[sub]
        linked = sum(1 for r in rows if r.get("catalog_tools"))
        pri = priority.get(sub, "P3")
        lines.append(
            f"| {pri} | {sub} | {len(rows)} | indexed | {linked} with catalog | mirror | batch {batch} |"
        )
        if pri == "P1":
            batch += 1
    matrix_path = root / args.matrix_out
    matrix_path.parent.mkdir(parents=True, exist_ok=True)
    matrix_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"Wrote matrix -> {matrix_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
