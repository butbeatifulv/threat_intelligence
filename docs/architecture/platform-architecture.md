# Veil platform architecture (current + target)

**Entry:** layer summary and quick start in [README.md](../README.md). **Ports:** [threatintel-runtime.md](../architecture/threatintel-runtime.md).

**Current runtime (2026-05):** three isolated Go modules in veil — `discovery/`, `pipeline/`, `knowledge/` — plus shared `pkg/*`. Pentest execution moved to **[veneno](https://github.com/butbeautifulv/veneno)**. Integration: NATS (`harvest` / `commit` / `engage.events` bridge) and HTTP (veneno → veil-api only).

**v8 (done):** logical layers **Discovery**, **Pipeline**, **Knowledge**, **Engage**, **Report**; shared `pkg/report`, `pkg/decision`, `pkg/exec`, `pkg/api`, `pkg/mcp`; browser in `discovery/pkg/browser`.

---

## Current state (what is done)

| Track | Status | Proof |
|-------|--------|--------|
| Platform P0–P4b | Bus tests, closed/full loop, Terraform skeleton | [platform-full-loop-smoke.md](platform-full-loop-smoke.md) |
| Platform P5 | Hybrid deploy skeleton | [deploy-platform-hybrid.md](../deploy/deploy-platform-hybrid.md) |
| Platform P6 | Infra DRY (events, auth, scrapepub, stacks, natsjet publish) | [archive/veil_platform_refactor_p6.plan.md](../.cursor/plans/archive/veil_platform_refactor_p6.plan.md) |
| Platform P7 | Tests + `pkg/*/domain` SOT + `make test-platform-p7` CI | [domain-contour.md](../architecture/domain-contour.md), [archive/veil_platform_p7_tests_then_pkg_domain.plan.md](../.cursor/plans/archive/veil_platform_p7_tests_then_pkg_domain.plan.md) |
| Platform v8 | Layer renames + shared pkg extractions + discovery browser | [archive/veil_platform_v8_layers_master.plan.md](../.cursor/plans/archive/veil_platform_v8_layers_master.plan.md) |

**Veneno (pentest):** tool catalog and exec deploy from [veneno](https://github.com/butbeautifulv/veneno). Veil ingests results via `pipeline/engage-events/` — [veil-veneno-split.md](veil-veneno-split.md).

---

## Target logical layers (v8)

These are **roles**, not necessarily one repo folder each. Go modules stay isolated; shared logic moves to `pkg/`.

```mermaid
flowchart TB
  subgraph clients [External clients]
    AG[Agents / operators]
  end
  subgraph edge [Unified edge - P12]
    NGX[nginx or platform/gateway TLS :443]
  end
  subgraph discovery [Discovery]
    SC[discovery / harvest]
    BR[browser crawl - discovery]
  end
  subgraph knowledge [Knowledge]
    GR[knowledge ingest + Neo4j]
    GS[veil-api /v1]
    GM[veil-mcp /mcp/graph]
    DE[decision / intel engine]
  end
  subgraph pipeline [Pipeline NED]
    NED[normalize enrich dedup]
  end
  subgraph engage [Engage]
    EA[engage-api /api]
    EM[engage-mcp /mcp/engage]
    TOOLS[pentest tool catalog + runner]
  end
  subgraph report [Report - shared]
    REP[PDF/HTML/executive templates]
  end
  subgraph exec [Execution plane - optional shared]
    RUN[runner + sandbox]
  end
  AG --> NGX
  NGX -->|/v1/*| GS
  NGX -->|/mcp/graph| GM
  NGX -->|/api/*| EA
  NGX -->|/mcp/engage| EM
  AG -.->|stdio MCP local| GM
  AG -.->|stdio MCP local| EM
  EA --> TOOLS
  EM --> TOOLS
  TOOLS --> RUN
  SC -->|harvest| NED
  BR -.->|future harvest| NED
  NED -->|commit| GR
  SC -.->|optional| RUN
  TOOLS -->|HTTP veil-api| GS
  TOOLS --> DE
  TOOLS --> REP
  GR --> GS
  DE --> GS
```

**Unified path contract (P12):** [platform-unified-access.md](../architecture/platform-unified-access.md). Today dev may still hit `:8090` / `:8890` directly; production targets a single TLS host with prefix routing.

| Layer | Responsibility | Path today | Path target |
|-------|----------------|------------|-------------|
| **Discovery** | Fetch raw intel; ledger; optional browser | `discovery/` | `discovery/` (**P8h done**) |
| **Pipeline** | Normalize, enrich, dedup | `pipeline/` | `pipeline/` |
| **Knowledge** | Neo4j ingest + read API + reasoning | `knowledge/` | `knowledge/` (**P8i done**) |
| **Engage (veneno)** | Pentest catalog, runner, guard | [veneno](https://github.com/butbeautifulv/veneno) | veneno repo (split from veil) |
| **Report** | HTML/PDF/executive | **`pkg/report`** | **`pkg/report`** (**P8b done**) |
| **API + MCP** | Agent HTTP/MCP façade | **`pkg/api`**, **`pkg/mcp`** | **`pkg/api`**, **`pkg/mcp`** (**P8d done**) |

**Hard rules:** no cross-import between runtime modules (`discovery`, `pipeline`, `knowledge`, `engage` after rename). NATS wire and `pkg/harvest` / `pkg/commit` **schemas unchanged** in P8h/P8i (subjects may still say `scrape.>` / `ingest.>`). Engage → knowledge read path: HTTP **veil-api** only.

---

## Layer renames (P8h / P8i)

| Rename | Scope | Keep stable (compat) |
|--------|--------|----------------------|
| ~~`scrape/`~~ → **`discovery/`** (P8h **done**) | Go module path, `deploy/discovery/`, Makefile `test-discovery`, docs | NATS `scrape.>`; `pkg/harvest`; envelope `source` values; binary `scrape_worker` one release |
| `knowledge/` → **`knowledge/`** | Go module, `deploy/graph/`, Makefile `test-graph*` | NATS `ingest.>`; `GRAPH_PACK_VERSION`; Neo4j labels; URLs `/v1/*`; product names **veil-api**, **veil-mcp** |

**Order:** merge **P8h + P8i** to `main` before large P8b–g refactors (or rebase feature branches once). Details: [archive/veil_platform_v8_layers_master.plan.md](../.cursor/plans/archive/veil_platform_v8_layers_master.plan.md) § P8h, P8i.

**Docs:** use **Discovery** / **Knowledge** in prose immediately; link to legacy paths until rename lands.

---

## Runner vs factory (important)

They solve **different** problems today. Unifying the **name** without splitting concerns would blur security boundaries.

| | **Discovery `factory`** (today `discovery/harvest/internal/factory`) | **Engage `runner`** |
|--|----------------------|---------------------|
| **Purpose** | Register scheduled **sources**; inject `ScrapeDeps` (ledger, feeds, NATS publishers) | Execute **catalog tools** (subprocess) with audit, cache, target guard |
| **Unit of work** | `Source.Run(ctx, deps)` per feed (ti, vuln, ds, …) | `Runner.Run(ctx, toolName, args)` per tool invocation |
| **I/O** | HTTP/GitHub via `feeds.Client`; publish `harvest` | `docker exec` into **engage-runner** image (or local PATH) |
| **Isolation** | Trust boundary = egress + rate limits; **no** subprocess sandbox | **Sandbox** (`runner.Sandbox`): allowlisted binaries, timeouts, `ProcessTracker` |
| **Analogue** | Cron + plugin registry | CI job runner + container isolation |

**Recommendation:** use **`pkg/exec`** (P8e) as the cross-cutting execution plane, not rename factory to runner. Plain HTTP feeds stay on `feeds.Client`; optional subprocess spike: `discovery/pkg/execfetch` with `-tags discoveryexec` (see `pkg/exec/README.md`).

| `pkg/exec` capability | Engage (now) | Scrape (future) |
|----------------------|--------------|-----------------|
| `Sandbox` (docker/local) | yes | optional **discovery-fetcher** container for untrusted CLI (e.g. headless browser, `git` clone) |
| `Executor` interface | `runner.Executor` | thin wrapper for rare scrape subprocesses |
| Audit / timeout / allowlist | tool audit store | harvest job audit (optional) |

**Keep `factory`** as discovery orchestration (which sources, policies, NATS subjects). **Keep engage `Runner`** as security-tool orchestration. Share only **primitives** underneath.

**Browser (P8g):** `discovery/cmd/browser-agent` (Playwright) + `discovery/browser` HTTP service publishes `harvest` `source=browser` / `kind=scrape_browser_inspect_raw` on `scrape.browser.events`. Engage proxies `DISCOVERY_BROWSER_URL` for catalog/MCP tools.

---

## v8 extractions (done)

| Component | Was | Now |
|-----------|-----|-----|
| Decision / attack chain / tool selection | veneno intelligence layer | [pkg/decision](../pkg/decision/) + veneno wiring |
| Report generation | veneno report layer | [pkg/report](../pkg/report/) + veneno adapter |
| Browser automation | veneno browser (proxied) | [discovery/pkg/browser](../discovery/pkg/browser/), [discovery/browser](../discovery/browser/) |
| HTTP/MCP transport helpers | per-layer duplicates | [pkg/api](../pkg/api/), [pkg/mcp](../pkg/mcp/) |
| Execution plane | veneno runner | [pkg/exec](../pkg/exec/) + veneno adapter |
| Domain entities | scattered | [domain-contour.md](../architecture/domain-contour.md) (P7 + P8a) |

---

## Shared transports (`pkg/api` vs `pkg/mcp`) — P8d

Layer `serve` binaries keep **route tables and tool handlers**; shared wire plumbing lives under `pkg/*` only (no cross-import between discovery, pipeline, knowledge).

| Package | Responsibility | Used by |
|---------|----------------|---------|
| **`pkg/api`** | JSON responses, prod-safe `WriteError`, `RegisterHealth`, `PostJSON`, JWT middleware wrapper (delegates to `pkg/auth/httpmiddleware`) | `knowledge/serve`, veneno HTTP routers |
| **`pkg/mcp`** | JSON-RPC message types, stdio framing, streamable HTTP POST/SSE, `tools/list` payload helper, `tools/call` param parse, tool-call RBAC helper | `knowledge/serve` veil-mcp, veneno MCP |

Each layer passes its RBAC permission (`PermGraphRead` vs `PermEngageToolRun`) and registers domain routes on `http.ServeMux`; MCP `Server` types implement `mcp.Processor` for layer-specific tool catalogs.

---

## Verification commands (handoff)

```bash
make test-platform-p7      # pkg domain + bus slices
make test-pkg-domain
make test-engage-events-pipeline  # veneno → veil ingest bridge
make check-graph-version     # after ingest/schema changes
```

Pentest prod reference: [eval/results/veil-pentest-prod-latest.md](../eval/results/veil-pentest-prod-latest.md) (0 HIGH / 0 MEDIUM after hardening).
