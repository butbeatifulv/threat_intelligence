# Veil (Vulnerability Exploitation Intelligence Layer)

![Veil](docs/assets/veil.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Veil** is a Neo4j-backed threat-intelligence platform: **discovery → pipeline → knowledge** on NATS JetStream, with **veil-api** and **veil-mcp** for graph read.

**Pentest execution** moved to **[veneno](https://github.com/butbeautifulv/veneno)** (successor to the engage layer).

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Agents / AI:** [AGENTS.md](AGENTS.md) · **Split ADR:** [veil-veneno-split.md](docs/architecture/veil-veneno-split.md)

---

## Current state (2026-05-20)

| Area | Status | Details |
|------|--------|---------|
| **Platform** | P0–P12 + v8 **done** | Unified edge `veil-edge` — [platform-unified-access.md](docs/architecture/platform-unified-access.md); history in [.cursor/plans/archive/README.md](.cursor/plans/archive/README.md) |
| **Playbook domain** | **done** (read path) | **754** Anthropic cybersecurity skills in corpus; structured `procedures-index.json`, ontology API, graph seed props — [external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md) |
| **Engage sign-off** | Phase 30 **done** | Legacy HexStrike `:8888` decommissioned (architecture + routes) — [engage-audit-report.md](docs/engage/engage-audit-report.md). **Not** the same as 158/158 subprocess execution |
| **Active tracks** | In progress | **P9f** `make test-engage-executable-matrix`; plans: [engage_tools_full_coverage](.cursor/plans/engage_tools_full_coverage.plan.md), [client-native master](.cursor/plans/engage_mcp_client_native_execution_master.plan.md) |
| **Catalog** | **158** names | 151 legacy MCP + 8 engage bridge — [engage-tools.md](docs/engage/engage-tools.md) |
| **Host install (lab)** | Core47 **46/47** | `ghidra` manual — [engage-install-linux.md](docs/engage/engage-install-linux.md) |
| **Lab security** | Prod pentest **No-Go** | 2× high: open `/api/tools`, `/v1/categories` — [engage-lab-pentest.md](docs/engage/engage-lab-pentest.md), [engage-red-blue-bugs.md](docs/engage/engage-red-blue-bugs.md) |

---

## What you get

| Capability | Description |
|------------|-------------|
| **Threat graph** | Versioned [graph packs](docs/contracts/graph-pack.md), HTTP API (`/v1/*`), read-only MCP |
| **Cyber playbooks** | **754** procedure skills (mirror + index + structured steps); MITRE/CSF mappings; optional `HAS_PLAYBOOK` edges |
| **Ingestion bus** | Scrape → NED → ingest over NATS (`pkg/harvest`, `pkg/commit`) |
| **Engage toolkit** | **158** catalog tools · bridge + subprocess matrix · **client-native host PATH** by default |
| **Decision + playbook (DRY)** | Effectiveness in `pkg/decision`; skills add catalog **boost** and procedure hints, not a second matrix — [cyber-domain-model.md](docs/architecture/cyber-domain-model.md) |
| **Closed loop** | Tool runs → `engage.events` → graph (`EngageToolRun`, `EngageFinding`) |
| **Unified edge (P12)** | One TLS host: `/v1/*`, `/api/*`, `/mcp/graph`, `/mcp/engage`; scale **4 / 8 / 16** |
| **Agent-ready** | **veil-mcp** (read) + **veil-engage** (exec), Keycloak RBAC, GAIA eval harness |
| **Prod path** | Terraform + Ansible + Helm; [veil-controls](deploy/security/veil-controls.yaml) |

---

## Architecture

System diagram and layer roles: see mermaid in [docs/architecture/platform-architecture.md](docs/architecture/platform-architecture.md) (or the summary below).

| Layer | Path | Role | MCP |
|-------|------|------|-----|
| **Discovery** | [discovery/](discovery/) | Feeds, Vitess ledger, `harvest` publish | — |
| **Pipeline** | [pipeline/](pipeline/) | NED → `commit`; [engage-events](pipeline/engage-events/) | — |
| **Knowledge** | [knowledge/](knowledge/) | Neo4j ingest + [serve](knowledge/serve/) API/MCP | `veil-mcp` (read) |
| **Engage** | [engage/](engage/) | Catalog tools, workflows, reports | `veil-engage` (exec) |

**Shared `pkg/`:** [harvest](pkg/harvest/), [commit](pkg/commit/), [natsjet](pkg/natsjet/), [auth](pkg/auth/), [engage](pkg/engage/), [report](pkg/report/), [decision](pkg/decision/), [playbook](pkg/playbook/) (ontology, procedure, cataloglink), [exec](pkg/exec/), [api](pkg/api/), [mcp](pkg/mcp/). **No Go imports** across the four layer roots.

**Playbook stack (`pkg/playbook`):** committed [corpus/](corpus/anthropic-cybersecurity-skills/skills/) + [mappings](pkg/playbook/corpus/mappings/) → generated indexes → veil-api `/v1/playbooks/*` and MCP `playbook_*`. Execution stays on engage catalog + `pkg/decision`.

**Agents:** **veil-mcp** + **veil-engage** only — [mcp-agents.md](docs/agents/mcp-agents.md). Legacy HexStrike is reference-only — [external-hexstrike.md](docs/external/external-hexstrike.md).

**Contracts:** [ingest-contract.md](docs/contracts/ingest-contract.md) · [threatintel-runtime.md](docs/architecture/threatintel-runtime.md) (ports) · [engage-runtime.md](docs/engage/engage-runtime.md) · [deploy/](deploy/)

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

### Unified edge (graph + engage, TLS)

[platform-unified-access.md](docs/architecture/platform-unified-access.md) · `make test-platform-unified-edge`

### Engage (host PATH default)

```bash
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp
curl -sS http://localhost:8890/health | jq .
```

**Execution:** subprocesses on the **MCP host `PATH`** ([engage-mcp-topology.md](docs/engage/engage-mcp-topology.md)). Optional **runner** profile = docker-exec lab only — [deploy/engage/README.md](deploy/engage/README.md).

Install CLIs: [engage-install-linux.md](docs/engage/engage-install-linux.md) · Lab pentest notes: [engage-lab-pentest.md](docs/engage/engage-lab-pentest.md).

### Full scrape pipeline

```bash
./scripts/ops/compose-up-full.sh
./scripts/test/smoke-discovery-e2e.sh --up && ./scripts/test/smoke-discovery-e2e.sh
```

### MCP stdio (Cursor / Claude)

| Server | Launcher | Example config |
|--------|----------|----------------|
| veil-mcp | [run-veil-mcp.sh](scripts/mcp/run-veil-mcp.sh) | [cursor.mcp.json.example](examples/mcp/cursor.mcp.json.example) |
| veil-engage | [run-veil-engage.sh](scripts/mcp/run-veil-engage.sh) | [engage.stdio.json.example](examples/mcp/engage.stdio.json.example) |

**veil-mcp playbook tools:** `playbook_search`, `playbook_get`, `playbook_for_technique`, `playbook_procedure`, `playbook_recommend_tools`, `playbook_ontology_subdomains` — see [external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md).

---

## Documentation

Full index by category: **[docs/README.md](docs/README.md)**.

### Essential

| Document | Contents |
|----------|----------|
| [AGENTS.md](AGENTS.md) | Agent workflow, tests, core47 quick path |
| [docs/architecture/threatintel-runtime.md](docs/architecture/threatintel-runtime.md) | Compose, ports, NATS, bootstrap |
| [docs/agents/mcp-agents.md](docs/agents/mcp-agents.md) | veil-mcp + veil-engage setup |
| [docs/playbooks/external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md) | 754 skills, indexes, API/MCP, seed |
| [docs/architecture/cyber-domain-model.md](docs/architecture/cyber-domain-model.md) | Knowledge vs engage vs decision (DRY) |
| [docs/playbooks/playbook-import-matrix.md](docs/playbooks/playbook-import-matrix.md) | Subdomain migration tracker |
| [deploy/README.md](deploy/README.md) | Layer compose, scaling, smokes |
| [docs/engage/engage-tools.md](docs/engage/engage-tools.md) | Catalog KPIs, matrices, assessment API |
| [docs/engage/engage-lab-pentest.md](docs/engage/engage-lab-pentest.md) | Install + self-pentest + HexStrike lab results |

### Engage depth

[engage/README.md](engage/README.md) · [engage-runtime.md](docs/engage/engage-runtime.md) · [engage-install-linux.md](docs/engage/engage-install-linux.md) · [engage-hardening.md](docs/engage/engage-hardening.md) · [engage-audit-report.md](docs/engage/engage-audit-report.md) · [engage-red-blue-lab.md](docs/engage/engage-red-blue-lab.md)

### Reference

[platform-architecture.md](docs/architecture/platform-architecture.md) · [platform-closed-loop-pilot.md](docs/architecture/platform-closed-loop-pilot.md) · [deploy-platform-hybrid.md](docs/deploy/deploy-platform-hybrid.md) · [deploy-secure.md](docs/deploy/deploy-secure.md) · [domain-contour.md](docs/architecture/domain-contour.md) · [ontology-appsec.md](docs/architecture/ontology-appsec.md) · [agent-evaluation-gaia.md](docs/agents/agent-evaluation-gaia.md) · [external-security-frameworks.md](docs/external/external-security-frameworks.md)

**Layer READMEs:** [discovery/](discovery/README.md) · [pipeline/](pipeline/README.md) · [knowledge/](knowledge/README.md) · [scripts/](scripts/README.md)

---

## Tests

Full matrix: run from repo root. CI: [platform.yml](.github/workflows/platform.yml), [engage.yml](.github/workflows/engage.yml), [agent-eval.yml](.github/workflows/agent-eval.yml).

| Area | Commands |
|------|----------|
| **Shared / platform** | `make test-pkg-shared` · `make test-platform-p7` · `make test-platform-p0` · `make test-platform-unified-edge` · `make test-platform-closed-loop` · optional `make test-platform-full-loop` |
| **Layers** | `make test-discovery` · `make test-pipeline` · `make test-knowledge` · `make test-knowledge-serve` · `make test-graph-read-smoke` |
| **Playbook / corpus** | `make check-corpus-mappings` · `make check-skills-index` · `make check-procedures-index` · `cd pkg && go test ./playbook/...` |
| **Engage gates** | `make test-engage` · `make test-engage-parity` · `make test-engage-executable-matrix` (**P9f**) · `make test-engage-hardening` · `make test-engage-events-pipeline` |
| **Eval / deploy** | `make test-agent-eval-pilot` · `make deploy-helm-template` · `make deploy-ansible-check` |

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
