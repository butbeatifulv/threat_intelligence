# Legacy MCP parity checklist

Reference: [`refs/hexstrike-ai-master/`](../refs/hexstrike-ai-master/) (MIT, **reference-only — not runtime**). Parity scripts under `scripts/engage/` may read that tree; **`engage/` and `pkg/engage/` must not** import or embed `.external/hexstrike*` paths (`make test-engage-external-guard`, P11d).

| Area | Legacy reference | Veil engage |
|------|------------------|-------------|
| Behavioral parity | ad hoc Flask responses | Golden JSON tests (`go test ./internal/usecase/ctf/...` and `./internal/usecase/bugbounty/...`, `-run Golden`); `make test-engage-ctf` / `make test-engage-bugbounty` chain those packages then existing shell smokes |
| MCP tools | ~151 `@mcp.tool` | [catalog/tools.yaml](../engage/serve/catalog/tools.yaml) (**158** names: 151 legacy + 8 engage bridge) |
| Tool execution | legacy subprocess on host | KPIs and gates: [engage-tools.md](engage-tools.md) — `make test-engage-executable-matrix` (**P9f**), [engage-tools-na-matrix.md](engage-tools-na-matrix.md), [engage-executable-gaps.md](engage-executable-gaps.md) |
| HTTP API | Python server :8888 | `engage-api` :8890 |
| Auth | none | Keycloak + RBAC ([pkg/auth](../pkg/auth/)) |
| Graph context | none | `ENGAGE_VEIL_API_URL` client |

Regenerate catalog:

```bash
make catalog-engage
```

Enable tools for dev in [tools.live.yaml](../engage/serve/catalog/tools.live.yaml) (`enabled: true` when binary on PATH).

## API routes

| Route | Status |
|-------|--------|
| `GET /health` | implemented |
| `GET /api/tools`, `POST /api/tools/{name}` | implemented |
| `POST /api/intelligence/analyze-target` | implemented (HTTP/DNS heuristics + veil) |
| `POST /api/intelligence/select-tools` | implemented (`RankTools`, enabled filter; `stealth` / `comprehensive` objectives) |
| `POST /api/intelligence/create-attack-chain` | implemented (20+ `attack_patterns` + ranked fallback) |
| `POST /api/intelligence/comprehensive-api-audit` | implemented (discovery, schema, JWT, GraphQL phases) |
| `POST /api/intelligence/technology-detection` | implemented (`TechnologyStack` enum, 15 values) |
| `POST /api/intelligence/smart-scan` | implemented (`max_tools`, sync parallel or async jobs; `rate_limit_check`; `ENGAGE_MAX_PARALLEL`) |
| `POST /api/intelligence/assessment-report` | implemented (smart-scan + `summary_report` + `executive_summary` + findings) |
| `POST /api/intelligence/optimize-parameters` | implemented |
| `POST /api/bugbounty/*` workflows | implemented (phased: `phases[]`, `estimated_time`, `tools_count`; optional `execute`) |
| `POST /api/ctf/create-challenge-workflow` | implemented |
| `POST /api/ctf/auto-solve-challenge` | implemented |
| `POST /api/ctf/suggest-tools` | implemented |
| `POST /api/ctf/team-strategy` | implemented |
| `POST /api/ctf/cryptography-solver` | implemented |
| `POST /api/ctf/forensics-analyzer` | implemented |
| `POST /api/ctf/binary-analyzer` | implemented |
| `POST /api/vuln-intel/cve-monitor` | implemented (NVD 2.0 feed + exploitability analysis) |
| `POST /api/vuln-intel/exploit-generate` | implemented (deterministic templates, no LLM) |
| `POST /api/vuln-intel/cve-lookup` | implemented (single CVE + veil `vuln` enrichment) |
| `POST /api/visual/*` | implemented (`summary-report` + `executive_summary`; `export-report` → HTML/PDF) |
| `GET /api/visual/scan-progress/{id}` | implemented (smart-scan / chain poll) |
| `POST /api/browser/inspect` | implemented (Playwright sidecar enrich) |
| `GET /api/processes/dashboard` | implemented (progress + `system_load`) |
| `GET /api/audit/recent` | implemented (JSONL store, `ENGAGE_AUDIT_DIR`) |
| `GET /api/audit/export` | implemented (NDJSON; optional `?since=` RFC3339) |
| `GET /api/playbooks` | implemented (`playbooks/bugbounty.yaml`) |
| `POST /api/playbooks/{name}/run` | implemented (YAML → SmartScan / workflow) |
| `POST /api/intelligence/correlate-threat` | implemented (veil-graph search + `engage_context` / `related_cves` / `cve_details` when CVE service wired) |
| `POST /api/intelligence/discover-attack-chains` | implemented (`cve_attack_paths`, `cve_stages` when CVE indicators present) |
| `POST /api/intelligence/target-timeline` | implemented (audit + graph + correlation + merged `timeline[]`) |
| `GET /api/intelligence/target-timeline` | implemented (`?target=&limit=&include_graph=`) |
| `POST /api/intelligence/execute-attack-chain` | implemented (sequential default; `parallel: true` uses bounded pool) |
| Findings parsers | nuclei JSONL, nmap, **ffuf** JSON/lines, **sqlmap** injection blocks (`findings/parse.go`) |
| Benchmark script | `scripts/benchmark/engage-hexstrike-parity.sh` — `make test-engage-benchmark` |
| `POST /api/audit/export-webhook` | implemented (`ENGAGE_AUDIT_WEBHOOK_URL`) |
| `GET /metrics` | implemented when `ENGAGE_METRICS_ENABLED=1` (Prometheus) |
| `POST /api/jobs`, `GET /api/jobs/{id}` | implemented |
| `GET /api/cache/stats`, `POST /api/cache/clear` | implemented (TTL cache) |
| `GET /api/telemetry` | implemented (uptime, jobs, processes, cache) |
| `GET /api/processes/*`, `POST terminate/pause/resume` | implemented |
| `POST /api/command` | implemented (catalog allowlist; `ENGAGE_ALLOW_RAW_COMMAND=1` lab only; denied when `ENGAGE_ENV=prod` / `ENGAGE_DENY_RAW_COMMAND=1`) |
| `POST /api/files/create|modify|delete`, `GET /api/files/list` | implemented (`ENGAGE_FILES_DIR`) |
| `POST /api/payloads/generate` | implemented (buffer/cyclic/random → `ENGAGE_FILES_DIR`) |
| `GET /api/jobs`, `POST /api/jobs/{id}/cancel` | implemented |
| Job backends | `ENGAGE_JOBS_MODE`: `memory`, `file`, `redis` (`ENGAGE_REDIS_URL`), `nats` (`ENGAGE_NATS_URL`, JetStream) |
| Browser tools | `ENGAGE_BROWSER_URL` sidecar — Playwright inspect (forms, security_analysis, tech); MCP `browser_agent_inspect` in-process |
| Secure deploy | `compose.secure.yml` + nginx :8443; `ENGAGE_DENY_RAW_COMMAND=1`; `make test-engage-secure` (nightly CI) |
| CI veil-stack e2e | `.github/workflows/engage.yml` job `engage-veil-stack`; `make test-engage-veil-stack-ci` |
| Graph pack gate | `make check-graph-version` in engage CI when ingest paths change |

## Route parity matrix (audit 2026-05-16)

Full CSV: [engage-route-parity.csv](engage-route-parity.csv). CI: `make test-engage-route-parity`.

| Class | Count | Engage handling |
|-------|-------|-----------------|
| **implemented** | 59 | Explicit handlers in `router.go` + `router_error_handling.go` |
| **na_unified_tool** | 90 | `POST /api/tools/{name}` (catalog) replaces per-binary Flask routes |
| **na_out_of_scope** | 7 | Python REPL, AI test payload, process autoscale/pool (see audit report) |

Legacy vuln-intel extras: `attack-chains`, `threat-feeds`, `zero-day-research` → aliases in `registerVulnIntel`.  
Legacy `error-handling/*` → read-only diagnostics API (in-process recovery).

Audit summary: [engage-audit-report.md](engage-audit-report.md).

## MCP (veil-engage)

- stdio: `engage/serve/cmd/mcp` (LSP framing, `tools/list`, `tools/call`)
- optional HTTP: `ENGAGE_MCP_HTTP_ENABLED=1` on `:8892`
- **intelligence bridge (Phase 11+16):** `tools/call` for `category: intelligence` and names like `comprehensive_api_audit`, `analyze_target_intelligence`, `create_attack_chain_ai`, `target_timeline_intelligence` route to in-process handlers (not subprocess stubs)
- **CTF bridge (Phase 17):** `category: ctf` and `ctf_*` tools → `/api/ctf/*`; playbooks `ctf-web`, `ctf-pwn` in `playbooks/ctf.yaml`
- **CVE bridge (Phase 20):** `monitor_cve_feeds`, `generate_exploit_from_cve` → in-process NVD lookup + deterministic exploit templates (`/api/vuln-intel/*`)
- **Browser/visual (Phase 21):** `browser_agent_inspect`, `get_process_dashboard`, `get_live_dashboard` → in-process; `GET /api/visual/scan-progress/{id}`
- example: [examples/mcp/engage.stdio.json.example](../examples/mcp/engage.stdio.json.example)

## Decision engine (HexStrike `IntelligentDecisionEngine`)

| Capability | Status |
|------------|--------|
| Effectiveness tables (web/api/ip/cloud/binary) | Full port in `decision.go` + `effectiveness_data.go`; parity vs `.external` via `make test-engage-decision-parity` |
| `select_optimal_tools` objectives (quick/comprehensive/stealth) | `SelectToolsForTarget` |
| Technology-specific tools (WordPress→wpscan, PHP→nikto) | `appendTechSpecificTools` + CMS boost |
| `optimize_parameters` (15+ tools) | `parameter.go` + profile-aware `OptimizeParametersWithContext` |
| `analyze_target` (attack surface, confidence, DNS) | `profile.go`, `detect.go` |
| `create_attack_chain` (success_probability, time estimates) | `CreateAttackChain` |
| `IntelligentErrorHandler` | `recovery` package + bounded backoff in `tools.Run` |
| LLM | Out of scope |

## Phase 11 ops (optional)

| Feature | Env / path |
|---------|------------|
| Postgres audit | `ENGAGE_AUDIT_POSTGRES_URL`, `ENGAGE_AUDIT_RETENTION_DAYS` |
| Cross-layer events | `ENGAGE_EVENTS_NATS_ENABLED=1`, `ENGAGE_NATS_URL`, subjects `engage.events.audit` / `engage.events.finding`; pipeline `engage-events-worker` → `ingest.engage.tool_run` / `ingest.engage.finding`; graph ingest → Neo4j |
| Graph read (engage) | veil-api: `GET /v1/categories/engage/search?q=`, `GET /v1/categories/engage/context?q=` (findings + `MAY_RELATE_TO` CVEs), `GET /v1/nodes/{hostname}` for `EngageTarget`; MCP `target_timeline_intelligence`; smoke: `make test-graph-engage-category`, e2e: `make test-engage-veil-stack` (manual) / `make test-engage-veil-stack-ci` (CI) |
| PDF engine | `ENGAGE_PDF_ENGINE=gofpdf` (default) or `wkhtml` (requires `wkhtmltopdf`) |
| Playbooks | `ENGAGE_PLAYBOOKS_PATH` or `engage/serve/playbooks/bugbounty.yaml` |
| Keycloak e2e | `deploy/engage/compose.keycloak.yml`, `make test-engage-keycloak` (401 without JWT, 200 with token; nightly `engage-secure.yml`) |
| Metrics smoke | `make test-engage-metrics` |
