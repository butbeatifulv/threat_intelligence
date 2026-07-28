#!/usr/bin/env python3
"""Build playbook search artifacts: chunks JSONL, Bleve index, optional Qdrant vectors."""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import uuid
from datetime import datetime, timezone
from pathlib import Path

FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)
SECTION_RE = re.compile(r"^##\s+(.+)$", re.MULTILINE)
ATTACK_RE = re.compile(r"\bT\d{4}(?:\.\d{3})?\b", re.I)
MAX_CHUNK_CHARS = 2400
OVERLAP_CHARS = 200
SCHEMA_VERSION = 1


def repo_root() -> Path:
    here = Path(__file__).resolve()
    for parent in [here.parent, *here.parents]:
        if (parent / "versions.env").is_file() and (parent / "go.mod").exists():
            return parent
    return Path.cwd()


def load_skills_index(root: Path, rel: str) -> list[dict]:
    path = root / rel
    doc = json.loads(path.read_text(encoding="utf-8"))
    return doc.get("skills") or []


def strip_frontmatter(raw: str) -> tuple[str, str]:
    fm = FRONTMATTER_RE.match(raw)
    if not fm:
        return "", raw
    return fm.group(1), raw[fm.end():]


def split_body(body: str) -> list[tuple[str, str]]:
    """Return (section_title, section_text) pairs."""
    parts: list[tuple[str, str]] = []
    matches = list(SECTION_RE.finditer(body))
    if not matches:
        text = body.strip()
        if text:
            parts.append(("", text))
        return parts
    for i, m in enumerate(matches):
        title = m.group(1).strip()
        start = m.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(body)
        text = body[start:end].strip()
        if text:
            parts.append((title, text))
    return parts


def chunk_text(text: str, max_chars: int, overlap: int) -> list[str]:
    text = text.strip()
    if not text:
        return []
    if len(text) <= max_chars:
        return [text]
    out: list[str] = []
    start = 0
    while start < len(text):
        end = min(len(text), start + max_chars)
        out.append(text[start:end])
        if end >= len(text):
            break
        start = max(0, end - overlap)
    return out


def build_chunks(root: Path, skills: list[dict], corpus_rel: str) -> list[dict]:
    chunks: list[dict] = []
    for skill in skills:
        sid = skill["id"]
        corpus_path = skill.get("corpus_path") or skill.get("external_path") or ""
        md_path = root / corpus_path
        if not md_path.is_file():
            continue
        raw = md_path.read_text(encoding="utf-8", errors="replace")
        _, body = strip_frontmatter(raw)
        attack_ids = skill.get("attack_ids") or []
        meta_text = " ".join(
            [
                skill.get("name") or sid,
                skill.get("description") or "",
                skill.get("subdomain") or "",
                " ".join(skill.get("tags") or []),
                " ".join(attack_ids),
            ]
        ).strip()
        chunks.append(
            {
                "chunk_id": f"{sid}::meta",
                "skill_id": sid,
                "subdomain": skill.get("subdomain") or "",
                "section_title": "_meta",
                "chunk_index": 0,
                "attack_ids": attack_ids,
                "text": meta_text,
                "kind": "meta",
            }
        )
        idx = 0
        for section_title, section_text in split_body(body):
            for piece in chunk_text(section_text, MAX_CHUNK_CHARS, OVERLAP_CHARS):
                idx += 1
                chunks.append(
                    {
                        "chunk_id": f"{sid}::{idx}",
                        "skill_id": sid,
                        "subdomain": skill.get("subdomain") or "",
                        "section_title": section_title,
                        "chunk_index": idx,
                        "attack_ids": attack_ids,
                        "text": piece,
                        "kind": "body",
                    }
                )
    return chunks


def write_chunks_jsonl(path: Path, chunks: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for row in chunks:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")


def build_bleve_index(chunks: list[dict], index_dir: Path) -> None:
    chunks_path = index_dir.parent / "playbook-chunks.jsonl"
    write_chunks_jsonl(chunks_path, chunks)
    import subprocess

    pkg_dir = repo_root() / "pkg"
    subprocess.run(
        ["go", "run", "./retrieval/cmd/indexbuild", str(chunks_path), str(index_dir)],
        check=True,
        cwd=pkg_dir,
        env={**os.environ, "GOWORK": "off"},
    )


def write_manifest(path: Path, chunk_count: int, skill_count: int, bleve_path: str) -> None:
    doc = {
        "schema_version": SCHEMA_VERSION,
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "chunk_count": chunk_count,
        "skill_count": skill_count,
        "bleve_index": bleve_path,
    }
    path.write_text(json.dumps(doc, indent=2) + "\n", encoding="utf-8")


def upsert_qdrant(chunks: list[dict], url: str, collection: str, embed_url: str, model: str, api_key: str) -> None:
    try:
        import urllib.request
    except ImportError:
        return
    # Minimal REST upsert — embeddings via Ollama/OpenAI HTTP
    points = []
    batch_size = 16
    dim = 384
    for i in range(0, len(chunks), batch_size):
        batch = chunks[i : i + batch_size]
        texts = [c["text"] for c in batch]
        vectors = embed_texts(texts, embed_url, model, api_key)
        if vectors and vectors[0]:
            dim = len(vectors[0])
            if i == 0:
                ensure_qdrant_collection(url, collection, dim)
        for row, vec in zip(batch, vectors):
            points.append(
                {
                    "id": str(uuid.uuid5(uuid.NAMESPACE_URL, row["chunk_id"])),
                    "vector": vec,
                    "payload": {
                        "chunk_id": row["chunk_id"],
                        "skill_id": row["skill_id"],
                        "subdomain": row.get("subdomain") or "",
                        "section_title": row.get("section_title") or "",
                        "text": row["text"][:2000],
                        "attack_ids": row.get("attack_ids") or [],
                    },
                }
            )
    ensure_qdrant_collection(url, collection, dim)
    body = json.dumps({"points": points}).encode("utf-8")
    req = urllib.request.Request(
        f"{url.rstrip('/')}/collections/{collection}/points?wait=true",
        data=body,
        headers={"Content-Type": "application/json"},
        method=PUT,
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        if resp.status >= 300:
            raise RuntimeError(f"qdrant upsert failed: {resp.status}")


def ensure_qdrant_collection(url: str, collection: str, dim: int) -> None:
    import urllib.request

    payload = json.dumps({"vectors": {"size": dim, "distance": "Cosine"}}).encode("utf-8")
    req = urllib.request.Request(
        f"{url.rstrip('/')}/collections/{collection}",
        data=payload,
        headers={"Content-Type": "application/json"},
        method=PUT,
    )
    try:
        urllib.request.urlopen(req, timeout=30)
    except Exception:
        pass


def embed_texts(texts: list[str], embed_url: str, model: str, api_key: str) -> list[list[float]]:
    import urllib.request

    if embed_url.rstrip("/").endswith("11434") or "/api/embeddings" not in embed_url:
        url = embed_url.rstrip("/") + "/api/embeddings"
        body = json.dumps({"model": model, "prompt": texts[0] if len(texts) == 1 else texts}).encode("utf-8")
        req = urllib.request.Request(url, data=body, headers={"Content-Type": "application/json"})
        with urllib.request.urlopen(req, timeout=120) as resp:
            data = json.loads(resp.read().decode())
        if "embedding" in data:
            return [data["embedding"]]
        return data.get("embeddings") or []
    url = embed_url.rstrip("/") + "/v1/embeddings"
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"
    body = json.dumps({"model": model, "input": texts}).encode("utf-8")
    req = urllib.request.Request(url, data=body, headers=headers)
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = json.loads(resp.read().decode())
    return [item["embedding"] for item in sorted(data.get("data") or [], key=lambda x: x.get("index", 0))]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--skills-index", default="docs/skills-index/cyber-skills.json")
    ap.add_argument("--chunks-out", default="docs/skills-index/playbook-chunks.jsonl")
    ap.add_argument("--bleve-dir", default="docs/skills-index/playbook-search.bleve")
    ap.add_argument("--manifest", default="docs/skills-index/search-index-manifest.json")
    ap.add_argument("--chunks-only", action="store_true")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--vectors", action="store_true")
    ap.add_argument("--skip-vectors", action="store_true", default=True)
    ap.add_argument("--qdrant-url", default="http://127.0.0.1:6333")
    ap.add_argument("--qdrant-collection", default="veil_playbooks")
    ap.add_argument("--embed-url", default="http://127.0.0.1:11434")
    ap.add_argument("--embed-model", default="nomic-embed-text")
    ap.add_argument("--embed-api-key", default="")
    args = ap.parse_args()

    root = repo_root()
    skills = load_skills_index(root, args.skills_index)
    chunks = build_chunks(root, skills, "")
    chunks_path = root / args.chunks_out
    write_chunks_jsonl(chunks_path, chunks)

    if args.chunks_only:
        print(f"OK: {len(chunks)} chunks -> {chunks_path}")
        return 0

    bleve_dir = root / args.bleve_dir
    build_bleve_index(chunks, bleve_dir)
    manifest_path = root / args.manifest
    write_manifest(manifest_path, len(chunks), len(skills), args.bleve_dir)

    if args.vectors and not args.skip_vectors:
        upsert_qdrant(chunks, args.qdrant_url, args.qdrant_collection, args.embed_url, args.embed_model, args.embed_api_key)

    if args.check:
        if len(chunks) < 2000:
            print(f"ERROR: expected >=2000 chunks, got {len(chunks)}", file=sys.stderr)
            return 1
        print(f"OK: {len(chunks)} chunks, manifest updated")
        return 0

    build_extra_indexes(root, repo_root() / "pkg")
    print(f"OK: {len(chunks)} chunks, bleve -> {bleve_dir}")
    return 0


def build_extra_indexes(root: Path, pkg_dir: Path) -> None:
    import subprocess

    pairs = [
        ("ti-chunks.jsonl", "ti-search.bleve"),
        ("vuln-chunks.jsonl", "vuln-search.bleve"),
    ]
    for jsonl_name, bleve_name in pairs:
        path = root / "docs/skills-index" / jsonl_name
        if not path.is_file():
            continue
        out = root / "docs/skills-index" / bleve_name
        subprocess.run(
            ["go", "run", "./retrieval/cmd/indexbuild", str(path), str(out)],
            check=True,
            cwd=pkg_dir,
            env={**os.environ, "GOWORK": "off"},
        )


if __name__ == "__main__":
    raise SystemExit(main())
