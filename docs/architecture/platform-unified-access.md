# Platform unified access (ADR — P12)

**Status:** Implemented (P12 complete — merged on `main`). Post **veil/veneno split**, engage exec is **optional**: deploy tool execution from [veneno](https://github.com/butbeautifulv/veneno); veil graph read and unified edge routing remain in this repo.  
**Supersedes:** per-layer nginx on `:443` with path `/mcp` only (graph) and separate engage host ports (`:8890`, `:8892`) as the **operator-facing** contract.  
**Implementation:** P12a–j ([veil_platform_p12_unified_access.plan.md](../.cursor/plans/archive/veil_platform_p12_unified_access.plan.md)); CI smoke: `make test-platform-unified-edge`.

**Context:** After v8, graph read (`veil-api`, `veil-mcp`) and engage exec (`veil-engage` HTTP/MCP) share `pkg/api` and `pkg/mcp` but still expose **different URL prefixes and processes**. Agents and operators need one TLS edge, stable path names, and horizontal scale for stateless tiers without merging offensive execution into the graph MCP process.

---

## Decision

1. **Single TLS edge** (nginx or `platform/gateway`) terminates HTTPS and routes by **path prefix**, not by merging binaries.
2. **HTTP API contract** — graph remains **`/v1/*`**; engage remains **`/api/*`** (HexStrike parity).
3. **MCP HTTP contract** — disambiguate with **`/mcp/graph`** and **`/mcp/engage`** on the edge; upstream services may still listen on legacy `/mcp` internally (strip prefix at proxy).
4. **MCP stdio** — unchanged: two processes (`knowledge/serve/cmd/mcp`, `engage/serve/cmd/mcp`); no unified stdio mux.
5. **Stateless scale** — `VEIL_API_SCALE`, `VEIL_MCP_GRAPH_SCALE`, `VEIL_ENGAGE_API_SCALE`, `VEIL_MCP_ENGAGE_SCALE` (values `4` | `8` | `16`) drive Compose/Kubernetes replica counts (P12c).
6. **Neo4j** — production profile is **Neo4j Enterprise 3-core** cluster (P12d); single `neo4j:5` Compose service stays for dev/CI only.

**Hard rules (unchanged from v8):** no cross-import between `discovery/`, `pipeline/`, `knowledge/`, `engage/`. Engage → knowledge reads only via HTTP veil-api (`/v1/*` behind the edge). Tool execution only through engage (`/api/tools/*`, engage MCP).

---

## Path map (edge → upstream)

| Edge path (client) | Layer | Upstream service | Upstream path (today) | Notes |
|--------------------|-------|------------------|------------------------|--------|
| `GET /health` | Facade | composite or graph | `GET /health` on veil-api | P12g may aggregate graph + engage health |
| `/v1/*` | Knowledge | `veil-api` | same | Categories, nodes, engage context read |
| `/api/*` | Engage | `engage-api` | same | Tools, jobs, intelligence, workflows |
| `/mcp/graph` | Knowledge | `veil-mcp` HTTP | `/mcp` | Streamable HTTP + optional SSE; strip `/graph` |
| `/mcp/engage` | Engage | `engage-mcp` HTTP | `/mcp` | Tool catalog MCP; strip `/engage` |

**Default ports (dev, direct — no edge):**

| Service | Port | Base URL |
|---------|------|----------|
| veil-api | 8090 | `http://localhost:8090/v1/...` |
| veil-mcp HTTP | 8091 | `http://localhost:8091/mcp` |
| engage-api | 8890 | `http://localhost:8890/api/...` |
| engage-mcp HTTP | 8892 | `http://localhost:8892/mcp` |

**Unified edge (target, TLS :443):**

```text
https://veil.example/
  /v1/...           → veil-api:8090
  /api/...          → engage-api:8890
  /mcp/graph        → veil-mcp:8091/mcp
  /mcp/engage       → engage-mcp:8892/mcp
  /health           → veil-api or gateway composite
```

Legacy direct URLs and examples under `examples/mcp/*.example` remain valid until P12j updates agent docs.

---

## MCP transports

| Mode | Graph (read) | Engage (exec) | When |
|------|--------------|---------------|------|
| **stdio** | `veil-mcp` binary or `scripts/mcp/run-veil-mcp.sh` | `engage/serve/cmd/mcp` | Cursor/Claude local dev; CI `mcp-smoke.sh` |
| **HTTP** | Edge `/mcp/graph` → upstream `/mcp` | Edge `/mcp/engage` → upstream `/mcp` | Remote agents, pentest-prod, K8s without sidecar stdio |

**stdio rules (unchanged):** JSON-RPC only on **stdout**; logs on **stderr** (`pkg/mcp` framing). Do not combine graph tools and engage tools in one MCP process.

**HTTP rules:** Streamable HTTP POST (+ optional SSE) per [MCP spec](https://modelcontextprotocol.io/); auth via `Authorization: Bearer` when `AUTH_ENABLED=1` / `VEIL_REQUIRE_AUTH=1`. Graph tools require read RBAC (`veil-reader`); engage `tools/call` requires engage run permission.

**Post-split (veneno):** when tool execution runs in veneno, set `UNIFIED_MCP_ENGAGE_URL` to the veneno MCP upstream (e.g. `http://veneno-mcp:8892/mcp`). The platform MCP gateway and unified nginx edge proxy `/mcp/engage` to that URL when configured; graph-only deployments omit it.

**Agent config (target after P12f/j):**

```json
{
  "mcpServers": {
    "veil-graph": { "url": "https://veil.example/mcp/graph", "timeout": 300 },
    "veil-engage": { "url": "https://veil.example/mcp/engage", "timeout": 300 }
  }
}
```

---

## Scaling variables (stateless tiers)

Applied by `deploy/platform/compose-scale-veil.sh` and Helm values (P12c). **Not** applied to Neo4j, NATS, or `engage-runner`.

| Variable | Allowed | Default (dev) | Replicas scale |
|----------|---------|---------------|----------------|
| `VEIL_API_SCALE` | `4`, `8`, `16` | `1` | `veil-api` / knowledge `api` |
| `VEIL_MCP_GRAPH_SCALE` | `4`, `8`, `16` | `1` | `veil-mcp` HTTP |
| `VEIL_ENGAGE_API_SCALE` | `4`, `8`, `16` | `1` | `engage-api` |
| `VEIL_MCP_ENGAGE_SCALE` | `4`, `8`, `16` | `1` | `engage-mcp` HTTP |

nginx/gateway upstream blocks use Docker DNS or K8s service names with `least_conn` / round-robin. Session stickiness is **not** required (MCP and API handlers are stateless; Neo4j and job state live elsewhere).

---

## Neo4j cluster profile (data plane)

| Profile | Use | Topology |
|---------|-----|----------|
| **dev/CI** | `deploy/knowledge/compose.yml` | Single `neo4j:5` container, APOC, local Bolt `7687` |
| **prod (P12d)** | `deploy/platform/neo4j-enterprise/` | **3 core** Neo4j Enterprise; causal cluster; routing via `neo4j://` cluster discovery |

**Invariants across profiles:** `GRAPH_PACK_VERSION`, ingest via `pkg/commit`, labels and `/v1/*` routes unchanged. Ingest workers and `veil-api` use the same Bolt credentials from secrets; scale ingest separately (not via `VEIL_*_SCALE` above).

**Failover expectation:** lose 1 core → cluster remains writable (quorum 2/3). Backup/restore and pack import procedures stay in [graph-pack.md](../contracts/graph-pack.md).

---

## Secure stack (P12h)

Single HTTPS entry for graph read and engage exec when using [secure-unified](../deploy/stacks/secure-unified.yml). Nginx config: [deploy/platform/nginx/veil.conf](../deploy/platform/nginx/veil.conf).

Only **platform nginx** publishes a host port (`UNIFIED_NGINX_HTTPS_PORT`, default `443`). Internal services use the Docker network; Neo4j has no host publish in secure overlays.

Nginx forwards `Authorization: Bearer …` on every `proxy_pass` location. JWT validation and RBAC run in each Go service; the edge does not terminate OIDC.

### Keycloak roles by path

Use one realm (e.g. `veil`) and map AD/LDAP groups to realm or client roles. Veil services read roles from the access token (`realm_access.roles`, `resource_access.<client>.roles`).

| Path prefix | Required roles (when `RBAC_ENABLED=1`) | Env overrides |
|-------------|--------------------------------------|-----------------|
| `/v1/*`, `/mcp/graph` | `veil-reader` and/or `veil-admin` | `RBAC_ROLE_READER`, `RBAC_ROLE_ADMIN` |
| `/health` | None at API (public liveness) | — |
| `/api/*`, `/mcp/engage` | `veil-engage-runner` and/or `veil-engage-admin` | `RBAC_ROLE_ENGAGE_RUNNER`, `RBAC_ROLE_ENGAGE_ADMIN` |

Recommended realm roles: `veil-reader`, `veil-admin`, `veil-engage-runner`, `veil-engage-admin`. Grant `veil-reader` without `veil-engage-runner` for read-only TI access.

See [auth-keycloak.md](auth-keycloak.md) for issuer setup, token examples, and MCP stdio tokens.

### Bring-up

```bash
mkdir -p deploy/platform/nginx/certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deploy/platform/nginx/certs/tls.key \
  -out deploy/platform/nginx/certs/tls.crt \
  -subj '/CN=localhost'

set -a
source deploy/profiles/secure-graph.env
set +a

docker compose \
  -f deploy/knowledge/compose.yml \
  -f deploy/knowledge/compose.secure.yml \
  -f deploy/platform/compose.secure-unified.yml \
  --profile mcp \
  up -d --build
```

Optional veneno stack (tool exec): deploy from [veneno](https://github.com/butbeautifulv/veneno) and point `UNIFIED_MCP_ENGAGE_URL` at veneno MCP. See [veil-veneno-split.md](veil-veneno-split.md).

Verify host binding: only `443` (or `UNIFIED_NGINX_HTTPS_PORT`). Call graph paths with a reader JWT; call `/api/*` and `/mcp/engage` only when veneno is wired, with engage-runner roles.

---

## Non-goals (P12)

- Merging engage tool execution into `veil-mcp`.
- Changing engage HTTP route names away from `/api/*`.
- Changing graph read routes away from `/v1/*`.
- Replacing NATS harvest/commit subjects.

---

## Verification (full P12 program)

```bash
make test-platform-unified-edge   # P12i Docker smoke (unified nginx edge)
make test-platform-p7
make test-knowledge test-engage
```

---

## References

- [platform-architecture.md](../architecture/platform-architecture.md) — layer diagram with unified edge
- [mcp-agents.md](../agents/mcp-agents.md) — stdio setup (updated in P12j)
- [deploy/README.md](../deploy/README.md) — compose stacks
- [deploy/stacks/secure-unified.yml](../deploy/stacks/secure-unified.yml) — secure stack SSOT
- [veil_platform_p12_unified_access.plan.md](../.cursor/plans/archive/veil_platform_p12_unified_access.plan.md) — phase branches (archived)
