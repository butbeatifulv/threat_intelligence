# Veil (Vulnerability Exploitation Intelligence Layer)

![Veil](docs/assets/veil.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Veil** is a Neo4j-backed threat-intelligence platform: **discovery → pipeline → knowledge** on NATS JetStream, with **veil-api** and **veil-mcp** for graph read.

**Pentest execution** moved to **[veneno](https://github.com/butbeautifulv/veneno)** (successor to the engage layer).

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Agents / AI:** [AGENTS.md](AGENTS.md) · **Split ADR:** [veil-veneno-split.md](docs/architecture/veil-veneno-split.md)

---

## Current state (2026-06)

| Area | Status | Details |
|------|--------|---------|
| **Platform** | P0–P12 + v8 **done** | Unified edge `veil-edge` — [platform-unified-access.md](docs/architecture/platform-unified-access.md) |
| **Playbook domain** | **done** (read path) | **754** Anthropic cybersecurity skills — [external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md) |
| **Pentest execution** | **moved** | [veneno](https://github.com/butbeautifulv/veneno) — see [veil-veneno-split.md](docs/architecture/veil-veneno-split.md) |

---

## What you get

| Capability | Description |
|------------|-------------|
| **Threat graph** | Versioned [graph packs](docs/contracts/graph-pack.md), HTTP API (`/v1/*`), read-only MCP |
| **Cyber playbooks** | **754** procedure skills (mirror + index + structured steps); MITRE/CSF mappings; optional `HAS_PLAYBOOK` edges |
| **Ingestion bus** | Scrape → NED → ingest over NATS (`pkg/harvest`, `pkg/commit`) |
| **Veneno bridge** | `engage.events` from [veneno](https://github.com/butbeautifulv/veneno) → graph (`EngageToolRun`, `EngageFinding`) via `pipeline/engage-events/` |
| **Agent-ready** | **veil-mcp** (graph read); tool execution in veneno |
| **Prod path** | Terraform + Ansible + Helm; [veil-controls](deploy/security/veil-controls.yaml) |

---

## Architecture

System diagram and layer roles: see mermaid in [docs/architecture/platform-architecture.md](docs/architecture/platform-architecture.md) (or the summary below).

| Layer | Path | Role | MCP |
|-------|------|------|-----|
| **Discovery** | [discovery/](discovery/) | Feeds, Vitess ledger, `harvest` publish | — |
| **Pipeline** | [pipeline/](pipeline/) | NED → `commit`; [engage-events](pipeline/engage-events/) | — |
| **Knowledge** | [knowledge/](knowledge/) | Neo4j ingest + [serve](knowledge/serve/) API/MCP | `veil-mcp` (read) |

**Shared `pkg/`:** [harvest](pkg/harvest/), [commit](pkg/commit/), [natsjet](pkg/natsjet/), [auth](pkg/auth/), [engage/events](pkg/engage/events/) (wire types for veneno ingest), [playbook](pkg/playbook/). **No Go imports** across layer roots.

**Agents:** **veil-mcp** only for graph read — [mcp-agents.md](docs/agents/mcp-agents.md). Pentest execution: [veneno](https://github.com/butbeautifulv/veneno).

**Contracts:** [ingest-contract.md](docs/contracts/ingest-contract.md) · [threatintel-runtime.md](docs/architecture/threatintel-runtime.md) · [deploy/](deploy/)

---

## Quick start

Compose under [deploy/](deploy/); presets in [deploy/stacks/](deploy/stacks/).

### Graph only (Neo4j + API + optional MCP)

```bash
docker compose -f deploy/knowledge/compose.yml up -d --build
docker compose -f deploy/knowledge/compose.yml --profile mcp up -d --build mcp   # optional
export VEIL_REPO_ROOT="$(pwd)"   # required for playbook paths when API runs outside repo cwd
curl -sS http://localhost:8090/health
make test-graph-read-smoke
```

Pack version: [versions.env](versions.env) → `GRAPH_PACK_VERSION` (currently **v0.4.7**).

**Playbook smoke (no Neo4j required for index read):**

```bash
make skills-index procedures-index
curl -sS 'http://localhost:8090/v1/playbooks/search?q=forensics&limit=3'
curl -sS 'http://localhost:8090/v1/playbooks/acquiring-disk-image-with-dd-and-dcfldd/procedure'
```

**Graph playbook edges (Neo4j + ATT&CK ingested):**

```bash
cd knowledge/ingest && env GOWORK=../go.work go run ./cmd/playbook_seed
```

### Unified edge (graph + veneno upstream)

[platform-unified-access.md](docs/architecture/platform-unified-access.md) · pentest services run in [veneno](https://github.com/butbeautifulv/veneno)

### Full scrape pipeline

```bash
./scripts/ops/compose-up-full.sh
./scripts/test/smoke-discovery-e2e.sh --up && ./scripts/test/smoke-discovery-e2e.sh
```

### MCP stdio (Cursor / Claude)

| Server | Launcher | Example config |
|--------|----------|----------------|
| veil-mcp | [run-veil-mcp.sh](scripts/mcp/run-veil-mcp.sh) | [cursor.mcp.json.example](examples/mcp/cursor.mcp.json.example) |

Tool execution MCP: [veneno](https://github.com/butbeautifulv/veneno).

**veil-mcp playbook tools:** `playbook_search`, `playbook_get`, `playbook_for_technique`, `playbook_procedure`, `playbook_recommend_tools`, `playbook_ontology_subdomains` — see [external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md).

---

## Documentation

Full index by category: **[docs/README.md](docs/README.md)**.

### Essential

| Document | Contents |
|----------|----------|
| [AGENTS.md](AGENTS.md) | Agent workflow, tests, core47 quick path |
| [docs/architecture/threatintel-runtime.md](docs/architecture/threatintel-runtime.md) | Compose, ports, NATS, bootstrap |
| [docs/agents/mcp-agents.md](docs/agents/mcp-agents.md) | veil-mcp setup |
| [docs/playbooks/external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md) | 754 skills, indexes, API/MCP, seed |
| [docs/architecture/cyber-domain-model.md](docs/architecture/cyber-domain-model.md) | Knowledge vs decision (DRY) |
| [deploy/README.md](deploy/README.md) | Layer compose, scaling, smokes |
| [docs/engage/README.md](docs/engage/README.md) | Pentest docs moved to veneno |

### Reference

[platform-architecture.md](docs/architecture/platform-architecture.md) · [platform-closed-loop-pilot.md](docs/architecture/platform-closed-loop-pilot.md) · [deploy-platform-hybrid.md](docs/deploy/deploy-platform-hybrid.md) · [deploy-secure.md](docs/deploy/deploy-secure.md) · [domain-contour.md](docs/architecture/domain-contour.md) · [ontology-appsec.md](docs/architecture/ontology-appsec.md) · [agent-evaluation-gaia.md](docs/agents/agent-evaluation-gaia.md) · [external-security-frameworks.md](docs/external/external-security-frameworks.md)

**Layer READMEs:** [discovery/](discovery/README.md) · [pipeline/](pipeline/README.md) · [knowledge/](knowledge/README.md) · [scripts/](scripts/README.md)

---

## Tests

Full matrix: run from repo root. CI: [platform.yml](.github/workflows/platform.yml).

| Area | Commands |
|------|----------|
| **Layers** | `make test-discovery` · `make test-pipeline` · `make test-knowledge` · `make test-graph-read-smoke` |
| **Veneno bridge** | `make test-engage-events-pipeline` |
| **Playbook / corpus** | `make check-corpus-mappings` · `make check-skills-index` · `make check-procedures-index` |

PR minimum: see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Graph packs

[docs/contracts/graph-pack.md](docs/contracts/graph-pack.md) · default **v0.4.7** in [versions.env](versions.env).

```bash
make graph-pack-export   # Neo4j must be running
make graph-pack-build
```

---

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n)[0] AS label, count(*) AS c ORDER BY c DESC LIMIT 20;
MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
MATCH (t:AttackTechnique {id:'T1059.001'})-[:HAS_PLAYBOOK]->(s:CyberSkill) RETURN count(s) AS playbooks;
```
