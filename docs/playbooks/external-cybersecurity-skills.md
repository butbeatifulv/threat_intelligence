# Anthropic Cybersecurity Skills (reference playbooks)

**Community project** (not affiliated with Anthropic PBC). License per skill: typically **Apache-2.0** (see each `SKILL.md` frontmatter).

| Zone | Committed path |
|------|----------------|
| Framework mappings (Veil SOT) | [pkg/playbook/corpus/mappings/](../pkg/playbook/corpus/mappings/) |
| Skill procedures | [corpus/anthropic-cybersecurity-skills/skills/](../corpus/anthropic-cybersecurity-skills/skills/) |
| Dev import source (gitignored) | `.external/Anthropic-Cybersecurity-Skills-main/` |

Domain model: [cyber-domain-model.md](../architecture/cyber-domain-model.md). Master plan: [.cursor/plans/cyber_corpus_vendor_master.plan.md](../.cursor/plans/cyber_corpus_vendor_master.plan.md).

## What Veil uses it for

| Layer | Role |
|-------|------|
| **Knowledge / veil-mcp** | Read-only **playbook** category — search and fetch procedure text for agents |
| **Neo4j (optional)** | `(:CyberSkill)` linked to `(:AttackTechnique)` via `HAS_PLAYBOOK` after seed |
| **Engage** | Does **not** register skills as catalog tools; may surface playbook hints via veil-api (P2) |

This is **agent procedure knowledge** (when/how to investigate), not subprocess execution. Tool runs stay on **veil-engage** ([engage-tools.md](engage-tools.md)).

## Index (operational truth)

| Artifact | Purpose |
|----------|---------|
| [docs/skills-index/cyber-skills.json](../skills-index/cyber-skills.json) | Machine-readable metadata (generated) |
| [docs/skills-index/README.md](../skills-index/README.md) | Schema and stats |
| `make corpus-import` | Rsync from `.external/` into committed paths |
| `make check-corpus-mappings` | CI: mappings SOT present + valid Navigator JSON |
| `make skills-index` | Regenerate index from committed `corpus/.../skills` |
| `make check-skills-index` | CI: fail if index is stale |
| `make search-index` | Build Bleve playbook search index (+ optional Qdrant vectors) |
| `make check-search-index` | CI: fail if search index is stale |
| `make check-procedures-index` | CI: fail if procedures index is stale |

Bodies are read from disk using `corpus_path` in the index (under `corpus/anthropic-cybersecurity-skills/skills/`). Regenerate: `make corpus-import` then `make skills-index` and `make procedures-index`.

## MCP tools (veil-mcp)

| Tool | Purpose |
|------|---------|
| `playbook_search` | Hybrid search over playbook chunks (BM25 + optional vector RRF); returns `score`, `snippet`, `match_type` |
| `playbook_get` | Full markdown for one skill id |
| `playbook_for_technique` | Skills referencing MITRE id (e.g. `T1003.001`) |
| `playbook_procedure` | Structured steps (`WhenToUse`, `Steps`, tool mentions) |
| `playbook_recommend_tools` | Map mentions → engage catalog tool names |
| `playbook_ontology_subdomains` | Subdomain registry + CSF hints |

See [mcp-agents.md](../agents/mcp-agents.md).

## HTTP API

| Method | Path |
|--------|------|
| GET | `/v1/playbooks/search?q=&subdomain=&limit=` |
| GET | `/v1/playbooks/{id}` |
| GET | `/v1/playbooks/by-technique/{technique_id}` |
| GET | `/v1/playbooks/subdomains` |
| GET | `/v1/playbooks/framework/mitre-layer` |
| GET | `/v1/playbooks/framework/coverage` |
| GET | `/v1/playbooks/framework/docs` |
| GET | `/v1/playbooks/{id}/procedure` |
| GET | `/v1/playbooks/{id}/recommend-tools` |
| GET | `/v1/playbooks/ontology/subdomains` |
| GET | `/v1/playbooks/ontology/technique/{technique_id}/skills` |

Import tracker: [playbook-import-matrix.md](playbook-import-matrix.md).

## Graph seed (P1b)

```bash
make skills-index
# After Neo4j is up with ATT&CK techniques ingested:
go run ./knowledge/ingest/cmd/playbook_seed
```

Or import Cypher from `scripts/knowledge/playbook-seed.cypher` (generated with `--emit-cypher`).

## What we do not do

- Edit mirrored upstream prose in place (fork with new id or `veil_derived: true`)
- Treat `.external/` as operational truth (use committed `corpus/` + `pkg/playbook/corpus/`)
- Publish skills through NATS `harvest`/`commit`
- Add skills to Engage `tools.yaml`
- Replace official MITRE STIX ingest ([discovery/harvest/.../lola](../discovery/harvest/internal/sources/lola/))

## Related

- [external-security-frameworks.md](../external/external-security-frameworks.md) — JCSF/DAF/OWASP
- [domain-contour.md](../architecture/domain-contour.md) — `pkg/playbook/domain`
- Upstream [README](../.external/Anthropic-Cybersecurity-Skills-main/README.md), [ATTACK_COVERAGE](../.external/Anthropic-Cybersecurity-Skills-main/ATTACK_COVERAGE.md)
