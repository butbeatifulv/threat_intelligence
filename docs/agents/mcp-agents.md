# MCP for agents (Veil)

Veil exposes **two MCP servers** for agents — keep them separate. **Tool execution runs only through veil-engage** (Go `engage` layer). Do not point agents at legacy HexStrike MCP or a Flask HTTP API on **`:8888`** — see [migration runbook](#migration-runbook-hexstrike-flask-8888--veil-engage) below.

| MCP | Layer | Purpose |
|-----|-------|---------|
| **veil-mcp** | Graph read | Query Neo4j TI data (categories, nodes, search) |
| **veil-engage** | Engage exec | Run security tools from YAML catalog (~158 names); replaces Python `hexstrike_mcp.py` + `hexstrike_server.py` |

Do not merge offensive tool execution into the graph MCP process.

**Unified edge (P12):** operators and remote agents may use **one TLS hostname** with path-based routing. HTTP MCP uses **`/mcp/graph`** (graph read) and **`/mcp/engage`** (tool exec); stdio remains **two processes**. Full path map, scale variables, and Neo4j cluster profile: [platform-unified-access.md](../architecture/platform-unified-access.md).

---

## veil-mcp (graph read)

The **veil-mcp** server exposes read-only threat-intel graph tools over [MCP](https://modelcontextprotocol.io/) **stdio** (LSP-framed JSON-RPC) and optional **Streamable HTTP**. It works with Cursor, Claude Desktop, VS Code Copilot, Roo Code, Cline, and any MCP-compatible client.

## Stdio transport (important)

MCP on stdio uses **stdout only for JSON-RPC** (`Content-Length` framing). All logs go to **stderr** (`SetupMCPLogger`). Do not redirect stderr to `/dev/null` when debugging — only stdout must stay clean for the client.

Recommended launcher (builds `bin/mcp` if missing):

```bash
./scripts/mcp/run-veil-mcp.sh
```

## Prerequisites

1. Neo4j with graph data (compose bootstrap or ingest pipeline).
2. Binary (or use the launcher above):

```bash
cd knowledge/serve && env GOWORK=../go.work go build -o bin/mcp ./cmd/mcp
```

Default Bolt: `neo4j://localhost:7687`, user `neo4j`, password `neo4jpassword`.

## Protocol versions

Supported: `2024-11-05`, `2025-03-26`. On `initialize`, the server **echoes** the client's version when it is in that set; otherwise it returns `2024-11-05`. Streamable HTTP clients that omit a version may receive `2025-03-26`.

## Agent compatibility matrix

| Client | Config location | Schema | Example |
|--------|-----------------|--------|---------|
| **Cursor** | Project `.cursor/mcp.json` or Settings → MCP | `mcpServers` + `command` | [cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example) |
| **Claude Desktop** | `claude_desktop_config.json` | `mcpServers` | [claude-desktop.mcp.json.example](../examples/mcp/claude-desktop.mcp.json.example) |
| **VS Code Copilot** | `.vscode/mcp.json` | `servers` + `type: "stdio"` | [vscode-copilot.mcp.json.example](../examples/mcp/vscode-copilot.mcp.json.example) |
| **Roo Code / Cline** | Same as Cursor | `mcpServers` | [cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example) |
| **Generic stdio** | Varies | `mcpServers` | [mcp.json.example](../examples/mcp/mcp.json.example) |
| **HTTP MCP** | Client-dependent | `url` + `headers` | [mcp.remote.json.example](../examples/mcp/mcp.remote.json.example) |
| **Auth (Keycloak)** | stdio `env` | `MCP_ACCESS_TOKEN` | [mcp.auth.json.example](../examples/mcp/mcp.auth.json.example) |

Common fields (recommended):

- `command`: path to [scripts/mcp/run-veil-mcp.sh](../scripts/mcp/run-veil-mcp.sh)
- `timeout`: `300` (seconds) for heavy Neo4j queries
- `description`: note that tools are **read-only** — do not use `alwaysAllow` as for offensive automation MCPs

### Cursor

Copy [examples/mcp/cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example), fix the `command` path, reload MCP in Settings.

### Claude Desktop

macOS: `~/Library/Application Support/Claude/claude_desktop_config.json` — merge the `mcpServers` block from [claude-desktop.mcp.json.example](../examples/mcp/claude-desktop.mcp.json.example).

### VS Code Copilot

Use [vscode-copilot.mcp.json.example](../examples/mcp/vscode-copilot.mcp.json.example) in `.vscode/mcp.json` (`servers`, not `mcpServers`).

### Manual verification checklist

| Step | Cursor | Claude Desktop | VS Code Copilot |
|------|--------|----------------|-----------------|
| Server shows as connected | Expected | Expected | Expected |
| `tools/list` includes `ti_list_categories` | Expected | Expected | Expected |
| `ti_health` returns `neo4j_ok: true` | Expected | Expected | Expected |
| Automated in CI | — | — | — |

CI covers stdio via `./scripts/smoke/mcp-smoke.sh` and unit tests via `make test-graph-serve`.

## Environment

| Variable | Default | Role |
|----------|---------|------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j user |
| `NEO4J_PASS` | `neo4jpassword` | Neo4j password |
| `NEO4J_DB` | `neo4j` | Database name |
| `MCP_ENV` | `local` | Log level (`local` / `dev` / `prod`) — logs on **stderr** |
| `APP_VERSION` / `MCP_VERSION` | `0.4.2` | Reported server version |
| `AUTH_ENABLED` | `0` | Require Keycloak JWT for `tools/call` |
| `RBAC_ENABLED` | `0` | Require `veil-reader` / `veil-admin` roles |
| `KEYCLOAK_ISSUER` | — | Realm issuer (required if auth on) |
| `MCP_ACCESS_TOKEN` | — | Access token for stdio MCP when auth on |
| `MCP_HTTP_ENABLED` | `0` | Streamable HTTP on `MCP_HTTP_LISTEN` (default `:8091`) |
| `MCP_HTTP_LISTEN` | `:8091` | HTTP listen address |
| `MCP_HTTP_PATH` | `/mcp` | MCP endpoint path |
| `MCP_HTTP_PREFER_SSE` | `0` | Prefer SSE responses on POST |
| `MCP_HTTP_BIND_LOCAL` | `0` | Bind HTTP to `127.0.0.1` only |

Full auth setup: [auth-keycloak.md](auth-keycloak.md).

## Docker (stdio)

```bash
docker build -f deploy/knowledge/docker/mcp.Dockerfile -t veil-mcp .
docker run -i --rm --network host \
  -e NEO4J_URI=neo4j://localhost:7687 \
  -e NEO4J_USER=neo4j -e NEO4J_PASS=neo4jpassword \
  veil-mcp
```

## Tools (parity with HTTP API)

| Tool | HTTP equivalent |
|------|-----------------|
| `ti_list_categories` | `GET /v1/categories` |
| `ti_list_kinds_in_category` | `GET /v1/categories/{category}/kinds` |
| `ti_nodes_by_category` | `GET /v1/categories/{category}/nodes` |
| `ti_search_in_category` | `GET /v1/categories/{category}/search` |
| `ti_get_node` | `GET /v1/nodes/{id}` |
| `ti_neighbors` | `GET /v1/nodes/{id}/neighbors` |
| `ti_health` | connectivity + runtime |
| `ping` | MCP keepalive (empty result) |
| `playbook_search` | `GET /v1/playbooks/search` — hybrid BM25 + vector (RRF); optional `mode`, returns `score`/`snippet` |
| `playbook_get` | `GET /v1/playbooks/{id}` — full SKILL.md body (64KB cap) |
| `playbook_for_technique` | `GET /v1/playbooks/by-technique/{technique_id}` — index + optional `HAS_PLAYBOOK` graph |
| `playbook_framework` | `GET /v1/playbooks/framework/*` — MITRE Navigator layer, coverage, mapping docs |
| `playbook_subdomains` | `GET /v1/playbooks/subdomains` — 26-domain taxonomy counts |

Example (DFIR):

```json
{"name":"playbook_search","arguments":{"query":"disk imaging","limit":5}}
```

Then `playbook_get` with `id` from results (e.g. `acquiring-disk-image-with-dd-and-dcfldd`).

Regenerate: `make corpus-import` (dev, from `.external/`) then `make skills-index`. Mappings SOT: [pkg/playbook/corpus/mappings/](../pkg/playbook/corpus/mappings/). See [external-cybersecurity-skills.md](../playbooks/external-cybersecurity-skills.md), [cyber-domain-model.md](../architecture/cyber-domain-model.md).

Legacy (deprecated): `ti_list_kinds`, `ti_get_nodes_by_kind`, `ti_search`.

## Smoke test

```bash
./scripts/smoke/mcp-smoke.sh
```

Uses protocol `2024-11-05` (typical for Cursor/Claude clients).

## Authentication (optional)

When `AUTH_ENABLED=1`:

- Set `KEYCLOAK_ISSUER`, `RBAC_*` as in [auth-keycloak.md](auth-keycloak.md).
- Put a valid access token in `MCP_ACCESS_TOKEN` in the MCP server `env` block.
- `initialize` / `tools/list` stay open on stdio; **`tools/call`** requires a valid token and role.

## MCP Streamable HTTP/SSE

```bash
export MCP_HTTP_ENABLED=1
./scripts/mcp/run-veil-mcp.sh
```

**Direct (dev):** `POST http://localhost:8091/mcp`, `GET /health`.

**Unified edge (prod / remote):** `POST https://<veil-host>/mcp/graph` (nginx strips `/graph` to upstream `/mcp`). Engage: `POST https://<veil-host>/mcp/engage`. See [platform-unified-access.md](../architecture/platform-unified-access.md).

Remote client (direct port): [mcp.remote.json.example](../examples/mcp/mcp.remote.json.example).

Example unified HTTP config (after P12b edge is up):

```json
{
  "mcpServers": {
    "veil-graph": { "url": "https://veil.example/mcp/graph", "timeout": 300 },
    "veil-engage": { "url": "https://veil.example/mcp/engage", "timeout": 300 }
  }
}
```

## veil-engage (tool execution)

Engage MCP runs separately from graph read:

```bash
./scripts/mcp/run-veil-engage.sh
```

| Client | Example |
|--------|---------|
| Cursor / Cline | [engage.stdio.json.example](../examples/mcp/engage.stdio.json.example) |
| HTTP MCP | [engage.http.json.example](../examples/mcp/engage.http.json.example) |

- Server name: `veil-engage`
- Methods: `initialize`, `tools/list` (~158 catalog tools), `tools/call` → `POST /api/tools/{name}` equivalent
- Auth: `AuthorizeEngageMCP` + role `veil-engage-runner` when `AUTH_ENABLED=1`
- Logs on **stderr** (same stdio rule as veil-mcp)
- Supported execution model: **client-native** (`ENGAGE_EXECUTION_PROFILE=client-native`) — tools run on the same host where this MCP process runs. Install CLI binaries on that execution host per [engage-client-dependencies.md](engage-client-dependencies.md).

Compose: `deploy/engage/compose.yml` (`engage-mcp` on :8892). Runner overlay `compose.runner.yml` is legacy lab/CI only (`ENGAGE_EXECUTION_PROFILE=docker-exec`). Docs: [engage-runtime.md](engage-runtime.md), [engage-legacy-parity.md](engage-legacy-parity.md).

### Cross-layer workflow (engage scan → graph read)

When `ENGAGE_EVENTS_NATS_ENABLED=1` and the events bus is running, tool runs and findings are ingested into Neo4j as `EngageToolRun` / `EngageFinding` nodes (category **`engage`**).

1. Run a scan with **veil-engage** (`httpx_probe`, `smart-scan`, etc.).
2. Query results with **veil-mcp** or veil-api: `GET /v1/categories/engage/search?q=example.com`.
3. Structured subgraph (findings + linked CVEs): `GET /v1/categories/engage/context?q=example.com`.
4. Resolve ingested target by hostname: `GET /v1/nodes/example.com` (`EngageTarget.name`).
5. Unified agent read-back: engage-api `POST /api/intelligence/target-timeline` or MCP `target_timeline_intelligence` (audit + graph + correlation).
6. Optional: `correlate_threat_intelligence` merges TI/vuln/engage hits and `related_cves` when `ENGAGE_VEIL_API_URL` is set.
7. Bug bounty phased plan: `POST /api/bugbounty/reconnaissance-workflow` with `{"domain":"example.com"}` or MCP `bugbounty_reconnaissance_workflow` — returns `workflow.phases` before optional tool execution.
8. CTF: `POST /api/ctf/create-challenge-workflow` or MCP `ctf_create_challenge_workflow`; auto-solve with `ctf_auto_solve_challenge`; playbooks `ctf-web` / `ctf-pwn`.
9. CVE intel: MCP `monitor_cve_feeds` or `POST /api/vuln-intel/cve-monitor` → `correlate_threat_intelligence` with `CVE-…` indicators → `generate_exploit_from_cve` / `POST /api/vuln-intel/exploit-generate` for deterministic PoC templates (lab only).
10. Browser/visual: `browser_agent_inspect` (forms + security score) → `smart-scan` with `scan_id` → poll `GET /api/visual/scan-progress/{id}` → `assessment-report` for `executive_summary`.

Smoke: `make test-engage-events-pipeline` (Docker, includes Neo4j assert with `--profile graph-ingest`). CTF: `make test-engage-ctf`. CVE: `make test-engage-cve`. Browser: `make test-engage-browser`.

## Migration runbook (HexStrike / Flask `:8888` → veil-engage)

Use this sequence when retiring the MIT reference stack (`hexstrike_mcp.py` FastMCP stdio → `hexstrike_server.py` Flask on **`:8888`**).

### 1. Disable legacy MCP and HTTP backend

| Step | Action |
|------|--------|
| MCP client | Remove or disable any MCP server whose `command` runs `hexstrike_mcp.py` (or wrappers that spawn it). |
| Processes | Ensure **no** dependency on **`http://…:8888`** for agent tool calls — that port is **not** part of Veil. |
| Compose / scripts | Drop any compose service or systemd unit that starts the Python Flask tool server alongside Veil. |

### 2. Enable veil-engage (execution)

Recommended launcher from repo root (sets `ENGAGE_*` defaults and `GOWORK`):

```bash
./scripts/mcp/run-veil-engage.sh
```

**Cursor / VS-style clients:** merge the `veil-engage` stanza from [engage.stdio.json.example](../examples/mcp/engage.stdio.json.example) into project `.cursor/mcp.json` (or Settings → MCP). Adjust `cwd` and `ENGAGE_CATALOG_PATH` paths to your checkout.

Optional **Streamable HTTP** MCP on the engage MCP process (`POST` …`/mcp`): set `ENGAGE_MCP_HTTP_ENABLED=1` and tune `ENGAGE_MCP_HTTP_LISTEN` / `ENGAGE_MCP_HTTP_PATH`; default listen **`:8892`** in config (see Compose `engage-mcp`). On the **unified edge**, use `https://<veil-host>/mcp/engage` instead of `:8892`. Do not confuse this with legacy **`:8888`**.

Default/Recommended: run `veil-engage` on the analyst host via [`scripts/mcp/run-veil-engage.sh`](../scripts/mcp/run-veil-engage.sh), install tools on that host, and keep graph read as a separate MCP. Compose runner overlay is optional legacy lab/CI path only — details in [engage-runtime.md](engage-runtime.md).

### 3. Keep graph read separate (unchanged)

After migration you still configure **veil-mcp** for Neo4j TI tools (same as before). Removing HexStrike **does not** collapse graph + exec into one process.

### 4. Environment quick reference (veil-engage MCP)

| Variable | Typical role |
|----------|----------------|
| `ENGAGE_ENV` | Log / behavior profile (`local` / `dev` / `prod`) |
| `ENGAGE_CATALOG_PATH` | Absolute or repo-relative path to `engage/serve/catalog/tools.yaml` |
| `AUTH_ENABLED` / `ENGAGE_AUTH_ENABLED` | `0` local; `1` with Keycloak per [auth-keycloak.md](auth-keycloak.md); MCP role **`veil-engage-runner`** when auth on |
| `ENGAGE_VEIL_API_URL` | Base URL for correlate / intelligence features that call veil-api |

Full engage API/runtime variables: [engage-runtime.md](engage-runtime.md).

### 5. Verify

| Check | Expected |
|-------|-----------|
| `tools/list` on **veil-engage** | ~158 catalog tools (parity: `make test-engage-parity`) |
| Legacy `:8888` | Nothing listening / not referenced in MCP config |
| **`make test-engage`** | Green in CI/local after any automation changes |

## Related

- [platform-unified-access.md](../architecture/platform-unified-access.md) — single TLS hostname, `/v1` + `/api` + MCP paths, scale 4/8/16
- [engage-runtime.md](engage-runtime.md) — engage API, runner modes, ports
- [external-hexstrike.md](../external/external-hexstrike.md) — MIT reference in `.external/` (superseded by engage layer)
- [auth-keycloak.md](auth-keycloak.md) — Keycloak, RBAC
- [deploy-secure.md](../deploy/deploy-secure.md) — production hardening
- [threatintel-runtime.md](../architecture/threatintel-runtime.md) — compose, ports
