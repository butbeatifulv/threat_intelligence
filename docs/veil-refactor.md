# Veil — refactor audit

**Date:** 2026-07-14
**Scope:** `projects/veil` — Go layers (`discovery/`, `pipeline/`, `knowledge/`, `platform/`, shared `pkg/`); goroutine/concurrency safety; DDD layering vs [`docs/agents/coding-style.md`](agents/coding-style.md); doc-vs-code drift after the veil/veneno split ([`veil-veneno-split.md`](architecture/veil-veneno-split.md))
**Method:** three parallel deep-read passes (not grep-only) — concurrency audit of all `go func` sites + NATS consumer loops, layering audit against the repo's own architecture rules, and a claims audit comparing README/docs status tables against actual code paths. Followed up with manual verification of every P0 finding.
**Audience:** this doc is the backlog for a refactor pass to be executed by **Cursor / Composer 2.5**. Code locations below carry matching `TODO(veil-refactor <ID>)` / `FIXME(veil-refactor <ID>)` comments — grep `veil-refactor` to jump straight to each site. IDs are stable anchors (`#g01`, `#l02`, `#c01`, …) referenced from those comments.

**Execution plan:** [.cursor/plans/veil-refactor_master.plan.md](../.cursor/plans/veil-refactor_master.plan.md) — 36 sub-phases, small diff per PR.

---

## Executive summary

| Area | Status | Notes |
|------|--------|-------|
| Cross-layer import boundaries (`discovery`/`pipeline`/`knowledge` isolation) | **Green** | Zero violations found — the 19 separate `go.mod` files structurally prevent accidental cross-imports, and no module declares one as a dependency of another |
| Domain purity (no I/O in `domain/`) | **Green** | Zero `domain/` directories anywhere import `neo4j`, `net/http`, `nats`, or `bolt` |
| `knowledge/serve` isolation (no NATS/scrape) | **Green** | Confirmed clean |
| `cmd/` layering (wiring only) | **Yellow** | 1 violation ([L01](#l01)) — raw Cypher + orchestration logic in a `cmd/` binary |
| `usecase/` port discipline | **Yellow** | 1 direct driver import ([L02](#l02)) + a repeated `usecase`/`internal/feeds` boundary blur across 4 of 7 discovery sources ([L03](#l03)–[L06](#l06)) |
| Concurrency / goroutine safety | **Red** | Production NATS consumers can crash-loop on a single malformed message ([G01](#g01)); the `veil-mcp` stdio server skips DB/telemetry cleanup on **every** graceful shutdown ([G02](#g02)/[G03](#g03)) |
| Doc/code drift from the veil→veneno split | **Red** | The unified MCP gateway fails hard out of the box ([C02](#c02)); a playbook feature silently no-ops forever ([C01](#c01)); 6 docs still describe a removed `engage/` module with dead paths/targets ([C03](#c03)–[C07](#c07)) |
| God objects | **Green/Yellow** | No true god objects; one borderline 444-line multi-role struct ([O01](#o01)), low priority |
| CI enforcement of architecture rules | **Yellow** | No automated import-boundary linter (egregore has `scripts/verify_import_boundaries.py`; veil has none) — the Green results above hold *today* but nothing stops drift; consider porting that script |
| Existing `TODO`/`FIXME` comments before this audit | **None** | `grep -rn "TODO\|FIXME\|XXX\|HACK"` across all `*.go` returned 0 hits — this audit is the first code-level backlog markup in the repo |

**Totals:** 23 tracked findings — 6 concurrency (G-series), 9 layering/god-object (L/O-series), 9 doc-vs-reality (C-series). 8 are P0.

---

## Part 1 — Concurrency / goroutines (G-series)

All ~22 `go func` launch sites were read in full (15 production, 7 test-only — test sites excluded). Go is 1.25.0 everywhere, so the pre-1.22 loop-variable-capture bug does not apply. No goroutine-per-message pattern exists in the NATS consumers (batches are processed synchronously), so there is no unbounded-spawn risk — the risk is panic-safety and cancellation, not resource exhaustion.

| ID | File:Line | Problem | Why it matters | Severity | Effort | Fix |
|----|-----------|---------|-----------------|----------|--------|-----|
| <a name="g01"></a>G01 | `pkg/natsjet/pull.go:65` (marked) — handler used by `pipeline/ned/internal/consumer/consumer.go`, `knowledge/ingest/internal/ingest/consumer.go`, `pipeline/connector/nats/engage_consumer.go` | `RunPullLoop`'s call to the message handler has no `recover()`; handlers unmarshal/validate/apply untrusted, externally-scraped JSON (CVE/GHSA/semgrep/sigma/TI feeds) with incomplete defensive checks | A single malformed or adversarial message panics the goroutine. These loops run inside `errgroup.Go`, which does **not** recover panics, so the panic crashes the whole worker process. The message is never Ack'd/Nak'd, so JetStream redelivers it on restart — **a poison-pill message can crash-loop a consumer indefinitely** | **P0** | S | Wrap the `handle(...)` call with `defer func(){ if r:=recover(); r!=nil { ... treat as error, Nak, log } }()` so a panic becomes a Nak, not a crash |
| <a name="g02"></a>G02 | `pkg/mcp/framed.go:27` (marked) `FramedRW.Read` | Takes a `ctx` parameter but never uses it — `ReadString`/`ReadFull` block indefinitely on stdin; the caller's `select { <-ctx.Done(): return }` in `RunStdio` is only checked *between* reads, not during a blocked one | On SIGTERM/SIGINT, the stdio MCP server's read loop can hang forever if stdin sends no more bytes/EOF — a genuine shutdown-time goroutine leak, and the direct cause of G03 | **P1** | M | Make the read cancellable: run the blocked read in its own goroutine and `select` on it vs `ctx.Done()`, or close the underlying reader on cancellation |
| <a name="g03"></a>G03 | `knowledge/serve/cmd/mcp/main.go:105-108` (marked) | `go func(){ <-rootCtx.Done(); os.Exit(0) }()` races the real shutdown path: `defer c.Shutdown()`, `defer shutdown(shCtx)` (OTel flush), `defer stop()`. `os.Exit` never runs deferred functions | Because `Run()` can hang per G02, this goroutine wins the race on effectively every SIGTERM — **Neo4j session/driver cleanup and OTel span/metric flush are silently skipped on every graceful restart/deploy**, risking leaked DB connections and lost telemetry | **P0** | S (after G02) | Fix G02 first so `Run()` actually returns on cancellation, then remove the `os.Exit` escape hatch and let `main` fall through to its own deferred cleanup |
| <a name="g04"></a>G04 | `discovery/pkg/proxypool/proxypool.go:158-160` `Transport.RoundTrip` | `time.Sleep(time.Until(nextAfter))` blocks for up to the full cooldown (default 2m) + jitter with no reference to `req.Context()` | If every proxy is cooling down, an HTTP call sleeps the full cooldown even if the caller's context was canceled/timed out seconds earlier — wastes goroutines, defeats caller timeouts under load | P2 | S | `select { case <-req.Context().Done(): return nil, req.Context().Err(); case <-time.After(time.Until(nextAfter)): }` |
| <a name="g05"></a>G05 | `discovery/harvest/internal/sources/vuln/internal/usecase/scrape.go:133,140,149`; `.../exploits.go:68,248`; `.../lola/internal/usecase/scrape.go:240,253`; `.../sbom/internal/usecase/scrape.go:113,137`; `.../ti/internal/feeds/http_helpers.go:26` | Retry/backoff loops use bare `time.Sleep(backoff)` instead of selecting on `ctx.Done()` during the sleep | On worker shutdown mid-retry, cancellation is ignored for up to ~60s, delaying process exit past typical container/k8s grace periods (risk of SIGKILL mid-write). **`pkg/natsjet/pull.go` in the same repo already does this correctly** (`select { <-ctx.Done(); <-time.After(backoff) }`), so this is a real deviation, not unfamiliarity with the pattern | P1 | S (mechanical, repeat 5×) | Replace each `time.Sleep(backoff)` with the `select`-on-`ctx.Done()` pattern from `pkg/natsjet/pull.go` |
| <a name="g06"></a>G06 | `pkg/observability/metrics_server.go:37-42`, used by every `cmd/*/main.go` | The `ListenAndServe` goroutine's bind/serve error is only `log.Warn`'d, never surfaced to the caller (`StartMetricsServer` returns `void`) | If the metrics port is already bound (common in restart races) or misconfigured, every worker silently runs with **no `/metrics` endpoint and no operator-visible failure** — looks like "healthy but no data," not "down" | P2 | S | Have `StartMetricsServer` return an error channel or take a fail-fast callback |

### Clean goroutine sites (verified — do not flag for refactor)

- `platform/gateway/internal/gateway/health.go:72-111` `probeUpstreams` — bounded, `sync.WaitGroup` correctly `Add`/`Wait`'d, distinct pre-sized slice indices per goroutine, per-call context timeout.
- `platform/mcp-gateway/internal/aggregator/aggregator.go:96-134` `listTools` — proper `wg.Add(2)`/`wg.Wait()`, no shared-state race (each goroutine writes its own result struct). Note: the *error-handling* behavior of this exact function is flagged separately as [C02](#c02) — the goroutine mechanics are fine, the fan-in policy is not.
- `discovery/proxybroker/internal/broker/broker.go` — consistent `sync.Mutex` around `items`/`order`/`leases`/`next`; `healthPass` snapshots keys under lock, releases before network I/O, re-acquires only to write results — good short-critical-section discipline.
- HTTP graceful-shutdown pattern (`signal.NotifyContext` + one goroutine doing `<-ctx.Done(); srv.Shutdown(...)`) used identically and correctly in `platform/gateway/cmd/veil-gateway/main.go`, `platform/mcp-gateway/cmd/veil-mcp/main.go`, `discovery/browser/cmd/serve/main.go`, `knowledge/serve/cmd/api/main.go`, `discovery/proxybroker/cmd/main.go`.
- `pkg/natsjet/pull.go`'s own outer loop — checks `ctx.Done()` every iteration, uses the correct `select`-based backoff (the reference pattern G05 should copy).
- `errgroup.WithContext` wiring in `pipeline/ned/internal/components/runtime.go:87-91` and `knowledge/ingest/cmd/ingest_worker/main.go:83-87` — correct, though G01 still applies underneath since `errgroup` doesn't recover panics.

---

## Part 2 — Layering & god objects (L/O-series)

Checked against the repo's own hard rules in [`docs/agents/coding-style.md`](agents/coding-style.md): `cmd/` = wiring only; `domain/` = no I/O; `internal/repository/` = ports; `internal/usecase/` = orchestration (no raw HTTP/Cypher); `internal/feeds/` = outbound HTTP; shared entities live only in `pkg/*/domain`. The two hard invariants (no cross-layer imports, no I/O in `domain/`) are **fully clean** — every violation below is narrower.

| ID | File | Violation | Rule broken | Severity | Effort | Fix |
|----|------|-----------|-------------|----------|--------|-----|
| <a name="l01"></a>L01 | `knowledge/ingest/cmd/playbook_seed/main.go` (marked) | `main()` writes raw Cypher (`MERGE`/`MATCH`) and loops over skills/ATT&CK IDs doing linking logic directly | "`cmd/` = wiring only... no business logic, no HTTP/Cypher/per-source transform code" | **P0** | S | Extract into a `knowledge/ingest/internal/usecase` (e.g. `playbookseed`) behind a `repository.PlaybookSeedRepository` port; `main()` only parses env/flags and calls `usecase.Run(ctx)` |
| <a name="l02"></a>L02 | `knowledge/serve/internal/usecase/read.go` `Ping()` (marked) | Imports `neo4j-go-driver` directly and runs `tx.Run(ctx, "RETURN 1", nil)` — raw Cypher inside `internal/usecase` | usecase must orchestrate via repository/query ports, not hold a direct driver dependency | **P0** | S | Add `Ping(ctx) error` to the `query.ReadExecutor`/`query.Service` port in `knowledge/connector/query`; drop the driver import from usecase |
| <a name="l03"></a>L03 | `discovery/harvest/internal/sources/ti/internal/feeds/runner.go` (package `feeds`, 612 LOC). `type Runner struct` at line 27, constructed by `NewRunner` at line 37 (takes `repository.GraphRepository` directly). Per-feed methods and their repo-write call sites: `runKEV` (line 155) → `r.Repo.UpsertKEVVulnerability` (line 181); `runPTRSS` (line 200) → `r.Repo.UpsertReport` (line 242), `r.Repo.UpsertIOC` (line 246); `runURLhaus` (line 273) → `UpsertIOC` (line 322); `runThreatFoxAPI` (line 341) → `UpsertIOC` (line 392); `runThreatFoxExport` (line 401) → `UpsertIOC` (line 443); `runMalwareBazaar` (line 461) → `UpsertIOC` ×2 (lines 505, 513); `runFeodo` (line 523) → `UpsertIOC` (line 560); `runOpenPhish` (line 569) → `UpsertIOC` (line 605). Top-level orchestration entry: `Run` (line 92) | `Runner`, in a package literally named `feeds`, does HTTP fetch/retry **and** directly calls the repo-write methods above (11 call sites across 7 feed sources) — full orchestration living in the "outbound HTTP" package. `ti` has **no `internal/usecase` directory at all**, unlike its 6 sibling sources | `internal/feeds/` = outbound HTTP only; `internal/usecase/` = orchestration | P1 | M | Create `discovery/harvest/internal/sources/ti/internal/usecase/runner.go`; move `Runner`/`NewRunner`/`Run` and the 7 `run*` orchestration methods there (they keep calling `repository.GraphRepository`); leave only the raw fetch/parse helpers each `run*` calls in `internal/feeds`. Mirrors `coderules`/`nuclei` (see "already correct" below) |
| <a name="l04"></a>L04 | `discovery/harvest/internal/sources/lola/internal/usecase/scrape.go:38,44-74,220-267` | `ScraperUsecase` constructs `http.Client`, TLS transport, proxy-pool wiring, and hand-rolls retry/backoff HTTP fetch (`fetchBytesDirect`) directly in usecase — duplicates the `feeds.Client`/`feeds.FetchIfDue` abstraction it already imports | same rule as L03 | P1 | S | Move `fetchBytesDirect` + `http.Client`/proxy-pool construction into `internal/feeds`; usecase constructor takes a ready `*feeds.Client`, as `sbom` already does |
| <a name="l05"></a>L05 | `discovery/harvest/internal/sources/ds/internal/usecase/ingest.go:24,30-53` | Same pattern as L04: `Ingestor` builds `http.Client`/proxy pool inline in its constructor, HTTP retry logic embedded in usecase | same rule as L03 | P1 | S | Same fix as L04 |
| <a name="l06"></a>L06 | `discovery/harvest/internal/sources/vuln/internal/usecase/scrape.go:31,38-73,97-178` + `exploits.go` | `ScraperUsecase` builds `http.Client`/proxy pool and hand-rolls `downloadNVDPage` retry/backoff/429/pagination in usecase; also uses a runtime type-assertion (`u.repo.(interface{ PublishNVDPage(...) })`, scrape.go:217-222) to reach a method not on the declared port | same rule as L03; type-assertion also weakens the `VulnerabilityRepository` port contract | P1 | M | Same HTTP fix as L04; add `PublishNVDPage` to `VulnerabilityRepository` directly instead of type-asserting at the call site |
| <a name="l07"></a>L07 | `discovery/harvest/internal/sources/ds/internal/usecase/graphstore.go:6` | The `graphStore` port interface is defined inside `internal/usecase`, not `internal/repository` — `ds` has no `internal/repository` dir at all, unlike its 6 siblings | "`internal/repository/` = ports/interfaces" | P2 | S | Move the interface into a new `discovery/harvest/internal/sources/ds/internal/repository` package for consistency |
| <a name="l08"></a>L08 | `pipeline/pkg/nvd/parse/types.go:4` (`Vulnerability`), `:13` (`CVSS`), `:19` (`CPE`); `pipeline/pkg/nvd/map/map.go:7` (`Vulnerability`), `:16` (`CVSS`), `:22` (`CPE`), `:27` (`func FromNVD(p parse.Vulnerability) Vulnerability`); canonical types already at `pkg/vuln/domain/entity.go:4,14,20` | Both `parse` and `map` redefine `Vulnerability`/`CVSS`/`CPE` with identical fields/json-tags to each other **and** to the already-canonical `pkg/vuln/domain` types (used elsewhere by discovery/pipeline/knowledge). `map.FromNVD` (line 27) is a pure field-by-field copy between structurally identical types | "shared entities must live in `pkg/*/domain` — no duplicate `type X struct`" | P1 | S | Delete the 3 struct defs in both `parse/types.go` and `map/map.go`; import `pkg/vuln/domain.Vulnerability/CVSS/CPE` directly in both packages; delete `FromNVD` entirely (parse output becomes the domain type directly) unless a caller-side field-by-field diff reveals the raw-NVD-JSON shape genuinely diverges (it currently doesn't) |
| <a name="o01"></a>O01 | `discovery/proxybroker/internal/broker/broker.go` (444 LOC, 9 methods) | Single `Broker` owns HTTP routing (5 handlers), in-memory pool state, **and** the background health-check loop | Not a documented hard rule (proxybroker is a standalone tool) — general god-object smell, but code within is well-factored (see clean-goroutine list above) | P3 | M | Optional split into `httpapi` + `pool` + `health` if this service grows further; at 444 LOC with one cohesive "proxy broker" responsibility, low priority today |

### Large files checked and found legitimately cohesive (no action needed)

- `pkg/commit/envelope.go` (443 LOC) — one `Envelope` type + ~30 small per-event-kind idempotency-key builders; large because it enumerates the wire contract, not because it mixes concerns.
- `discovery/harvest/internal/feeds/github.go` (438 LOC) — free functions only, all GitHub API/codeload-zip fetch helpers, correctly placed in `internal/feeds`.
- `knowledge/ingest/internal/sources/{lola,ti}/storage/neo4j.go` (360/300 LOC) — `Store` adapters with 9-12 `Upsert*`/`Merge*` methods for one domain each, correctly in `storage/`.
- `knowledge/connector/query/service.go` (353 LOC) — one "read graph" responsibility, 9 query methods + private row-decoding helpers.
- `knowledge/serve/internal/transport/httpserver/router.go` (269 LOC) — pure route registration, correctly wiring-only.
- `discovery/harvest/internal/sources/{coderules,nuclei}/internal/usecase/*` and `.../sbom/internal/usecase/scrape.go` — these three show the **correct** pattern (usecase calls a separate `internal/feeds` package via `feeds.FetchIfDue`, no inline `http.Client{}`). **Use these as the template when fixing L03–L06.**

**Note on L03–L06:** this is one repeated anti-pattern across 4 of 7 `discovery/harvest` sources (`ti`, `lola`, `ds`, `vuln`) — HTTP client construction + proxy-pool wiring + retry/backoff duplicated almost verbatim four times instead of living once in `internal/feeds`. Templating the `coderules`/`nuclei`/`sbom` pattern out to the other four resolves L03–L06 together and collapses ~150 LOC of near-duplicate bootstrap code.

---

## Part 3 — Claimed vs. actual (C-series)

Context: veil recently split pentest execution out to a sister repo (**veneno**, per [`veil-veneno-split.md`](architecture/veil-veneno-split.md), 2026-06-23). The split is clean in code and `Makefile` (`engage/`, `deploy/engage/`, `scripts/mcp/run-veil-engage.sh` are gone; `Makefile` explicitly redirects engage targets to veneno). The problem is **docs and one live component still assume the old topology**. That mismatch is the dominant failure class below — verified README claims (754 skills, MCP tool registry, scrape-source factory) all check out and are **not** flagged.

| ID | Claim / location | Reality | Severity | Effort | Fix |
|----|-------------------|---------|----------|--------|-----|
| <a name="c01"></a>C01 | `playbook_recommend_tools` MCP tool + `GET /v1/playbooks/{id}/recommend-tools` (advertised in `docs/playbooks/external-cybersecurity-skills.md:44`) | `pkg/playbook/cataloglink/resolve.go` (marked) reads `engage/serve/catalog/tools.yaml` — a path removed by the split. `os.ReadFile` error is swallowed, so `catalogNames` stays empty forever and every call **silently returns nil**, with no error surfaced anywhere | **P0** | S | Point at a configurable veneno catalog path/URL, or drop the tool/endpoint from the advertised list until cross-repo sync exists; at minimum log a warning on 0 catalog files found |
| <a name="c02"></a>C02 | `docs/architecture/platform-unified-access.md:3` — unified MCP edge "Status: Implemented (P12 complete)" | `platform/mcp-gateway/internal/aggregator/aggregator.go` (marked) fails the **entire** `tools/list`/`tools/call` response if either graph or engage backend errors. `EngageMCPURL` defaults to `http://127.0.0.1:8892/mcp` with nothing listening there post-split. Out of the box, `veil-mcp`'s unified gateway errors out entirely instead of degrading to graph-only | **P0** | M | Make the engage backend optional — partial-result on failure (graph tools + warning) — or require operators to explicitly configure `UNIFIED_MCP_ENGAGE_URL` before the aggregator treats engage as present; document the new cross-repo dependency |
| <a name="c03"></a>C03 | `docs/architecture/platform-unified-access.md:149` (`-f deploy/engage/compose.yml \`), `:150` (`-f deploy/engage/compose.secure.yml \`) — bring-up snippet | Neither file exists (`deploy/` no longer has an `engage/` subdir; confirmed by directory listing). Copy-pasting this command fails immediately | **P0** (docs) | S | Delete those two `-f` lines from the snippet (or replace with a link to veneno's own compose bring-up), keep the rest of the knowledge-stack bring-up as-is |
| <a name="c04"></a>C04 | `docs/agents/mcp-agents.md` — dead references at lines **193** (`./scripts/mcp/run-veil-engage.sh`), **205** (link to `engage-client-dependencies.md`), **207** (`deploy/engage/compose.yml`, links to `engage-runtime.md`/`engage-legacy-parity.md`), **243** (`./scripts/mcp/run-veil-engage.sh` again), **250** (same script + `engage-runtime.md`), **265** (`engage-runtime.md`), **278** (`engage-runtime.md`). This spans the whole "veil-engage (tool execution)" section (~lines 188-283) | None of `scripts/mcp/run-veil-engage.sh`, `deploy/engage/compose.yml`, `engage-runtime.md`, `engage-client-dependencies.md`, `engage-legacy-parity.md` exist. Directly contradicts this repo's own `README.md:7` ("Pentest execution moved to veneno") and `docs/engage/README.md:3` ("Moved to veneno"). This is the doc `README.md` explicitly points agents to — an agent following it will try to run nonexistent scripts | **P0** (docs) | S | Delete the whole "veil-engage (tool execution)" section (~lines 188-283) from `mcp-agents.md`; replace with a one-line pointer to veneno's own agent docs, mirroring the pattern already used in `docs/engage/README.md:3` |
| <a name="c05"></a>C05 | `docs/architecture/platform-architecture.md` — line **5** ("four isolated Go modules — `discovery/`, `pipeline/`, `knowledge/`, `engage/`"), line **25** (`ENGAGE_CATALOG_PATH`/`compose.runner.yml` under `deploy/engage/`), line **93** (table row "Engage \| Pentest catalog, runner, guard \| `engage/` \| `engage/` (slim, **P8f done**)"), lines **144-148** (capability-map table citing `engage/.../intelligence`, `.../report`, `.../browser`, `.../runner`), lines **159-160** (`pkg/api`/`pkg/mcp` "used by" column listing `engage/serve`) | `engage/` does not exist at repo root — relocated to veneno per the ADR, but the doc still presents it as a present, "done" veil module in 6 separate places | P1 (docs) | S | Update all 6 locations: drop `engage/` from the module list (line 5), remove/relabel the P8f row (line 93) as moved-to-veneno, update the capability-map (144-148) and `pkg/api`/`pkg/mcp` consumer lists (159-160) to say "veneno" instead of `engage/serve` |
| <a name="c06"></a>C06 | `docs/architecture/platform-architecture.md` — line **15** ("HexStrike → Engage \| **Done** (Phases 16-30) \| [engage-audit-report.md](engage-audit-report.md)"), line **16** ("Tool catalog \| **158** names ... \| `make test-engage-parity`"), line **17** ("Executable matrix \| Partial → **158/158** target \| `make test-engage-executable-matrix` (**P9f**)"), line **171** (`make test-engage-parity    # 150 HexStrike names`) | `Makefile` no longer defines `test-engage-parity`/`test-engage-executable-matrix` (only the veneno-bridge `test-engage-events-pipeline` remains — confirmed via `grep "^test-" Makefile`). `docs/engage/engage-audit-report.md` doesn't exist (`docs/engage/` now contains only `README.md`). All 4 "proof" links/targets are dead | P1 (docs) | S | Remove or archive this whole status table (lines ~13-18 and the line-171 command example) — it documents veneno's history, not veil's current state; link out to veneno's own audit report if one exists there |
| <a name="c07"></a>C07 | `docs/deploy/deploy-platform-hybrid.md:7` ("Engage / HexStrike migration **signed off** — [engage-audit-report.md](engage-audit-report.md) Phase 30; operators use **veil-engage** only"); `docs/agents/agent-evaluation-gaia.md:12` (`| Security tools | make test-engage-parity | engage PRs |`) | Same dead targets/paths as C06 | P1 (docs) | S | `deploy-platform-hybrid.md:7`: rewrite to point at veneno, drop the dead link. `agent-evaluation-gaia.md:12`: remove the row or repoint it at veneno's own CI gate |
| <a name="c08"></a>C08 | `discovery/harvest/internal/factory/register.go:28` — `"scrape source %q is not implemented in scrape-worker yet"` | Verified as a guard for unknown/mistyped names, **not a real gap** — all 7 documented sources (`ds,vuln,lola,ti,sbom,coderules,nuclei`) are registered via `init()`. Branch is unreachable under any documented config | P2 | S | No functional fix; reword to "unknown scrape source" so it doesn't imply a documented-but-missing feature |
| <a name="c09"></a>C09 | `knowledge/ingest/internal/sources/vuln/storage/neo4j.go:46-48` — comment: "Minimal implementation for interface completeness... only returns whether the node exists" `FindByCVE` | Confirmed zero production callers (only referenced by the port interface and test fakes). Never joins CWE/CPE/Exploit relations despite the schema defining them — genuinely incomplete, but currently dead code | P2 | S/M | Wire into a real CVE-lookup read path with full relations, or remove the method if truly unused, so a future caller doesn't get a silently truncated result |

### Verified accurate (no finding — checked, not assumed)

- README's "754 skills" claim: `corpus/anthropic-cybersecurity-skills/skills/` has exactly 754 dirs; `docs/skills-index/cyber-skills.json` has exactly 754 entries.
- `veil-mcp`'s tool registry (`knowledge/serve/internal/transport/mcpserver/tools.go`) matches `mcp-agents.md`'s graph-read tool table.
- The `engage` read-only category (`knowledge/connector/query/categories.go:53-68`) is real and queryable — the `pipeline/engage-events/` ingestion bridge itself is not in question, only the doc/tooling drift above.

---

## Detailed fix guidance for the harder items

The mechanical fixes (add a port method, delete a doc line, wrap a call in `recover()`) are self-explanatory from the tables above. These four need a design sketch because there's more than one way to do them and the wrong choice re-introduces the bug in a different shape.

### G01 — panic-safe message handling (`pkg/natsjet/pull.go:65`)

Wrap the handler call itself, not the outer loop (wrapping the loop would stop the whole consumer on one panic instead of just Nak'ing one message):

```go
func() {
    defer func() {
        if r := recover(); r != nil {
            log.Warn("message handler panicked", slog.Any("panic", r))
            if opts.NakDelay > 0 {
                _ = m.NakWithDelay(opts.NakDelay)
            } else {
                _ = m.Nak()
            }
        }
    }()
    if err := handle(msgCtx, m); err != nil {
        // existing Nak-on-error path
    }
}()
```
Note the existing `if err != nil { ...; continue }` logic needs to move inside this closure, and the closure needs a way to signal "already handled, skip the Ack" back to the outer loop (e.g. a small `handled bool` return) since `continue` inside the inner closure won't affect the outer `for` loop.

### G02 — cancellable stdio read (`pkg/mcp/framed.go:27`)

The blocking call is `bufio.Reader.ReadString`/`io.ReadFull` on `os.Stdin`, which Go cannot interrupt from another goroutine directly. The standard pattern is to run the blocking read in its own goroutine and select on a done-channel vs `ctx.Done()`:

```go
func (rw *FramedRW) Read(ctx context.Context) ([]byte, error) {
    type result struct {
        buf []byte
        err error
    }
    ch := make(chan result, 1)
    go func() {
        buf, err := rw.readFrame() // existing body, extracted unchanged
        ch <- result{buf, err}
    }()
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case r := <-ch:
        return r.buf, r.err
    }
}
```
Caveat: the spawned goroutine leaks until stdin actually produces a byte or EOF (Go has no portable "close stdin from another goroutine" primitive without touching the fd directly) — this stops the *process* from hanging on shutdown (which is the actual bug, G03's root cause), but the leaked goroutine itself is harmless since the process exits right after. Document that tradeoff in the fix's commit message rather than trying to eliminate the leak with fd tricks — not worth the platform-specific complexity here.

### G03 — remove the `os.Exit` shutdown race (`knowledge/serve/cmd/mcp/main.go:105`)

Once G02 makes `Run()` return promptly on `ctx.Done()`, delete the `go func(){ <-rootCtx.Done(); os.Exit(0) }()` block entirely (lines 105-108) and let control fall through normally: `Run()` returns nil on clean cancellation → `main()` reaches its end → deferred `c.Shutdown()`, OTel `shutdown()`, and `stop()` all run in order. If `Run()`'s cancellation-path error needs distinguishing from a real error (so `main` doesn't log a scary "mcp server stopped" error on a normal SIGTERM), check `errors.Is(err, context.Canceled)` before the `logger.Error`/`os.Exit(1)` branch at line 110-113.

### C02 — degrade-gracefully aggregator (`platform/mcp-gateway/internal/aggregator/aggregator.go:96`)

Two independent decisions, both needed:
1. **`listTools`**: when one backend errors, log it and return the other backend's tools instead of failing the whole call — i.e. only `return nil, err` if **both** `graphListed.err != nil && engageListed.err != nil`. When engage fails, log at `Warn` (not `Error` — this will be the common case until an operator configures a veneno MCP URL) and continue with `merged = graphListed.tools` only.
2. **`routeTool`** (the `tools/call` dispatcher, ~line 183 onward): a tool call *addressed to* the engage backend when engage is unreachable should return a clear JSON-RPC error naming the missing backend (e.g. "engage backend not configured — see UNIFIED_MCP_ENGAGE_URL"), not a generic connection-refused error, so a calling agent can distinguish "this tool doesn't exist" from "this tool exists but its backend isn't wired up."

Also update `platform/mcp-gateway/internal/config/config.go`'s `EngageMCPURL` default: keeping `http://127.0.0.1:8892/mcp` as a default is fine (it documents the expected port) as long as (1) above stops that unreachable default from being fatal, and `docs/architecture/platform-unified-access.md` gets a note that engage is optional post-split ([C02](#c02) pairs with the doc fix in [C03](#c03)).

### L03–L06 — discovery usecase/feeds template

The 4 broken sources (`ti`, `lola`, `ds`, `vuln`) should end up structured like `coderules` (already correct, confirmed by reading it):

- `internal/feeds/` — only `feeds.Client`, `feeds.FetchIfDue(ctx, client, ledger, key, source, url, policy, cachePath, buildReq)`, `feeds.GitHubRefs/GitHubRawURL/GitHubFetchRaw` — i.e. **only** the generic fetch/cache/ledger primitives and raw HTTP helpers, no per-source business logic, no direct repo-writing.
- `internal/usecase/runner.go` — `type Runner struct { feeds *feeds.Client; ledger *ledger.Store; repo <port> ; log *slog.Logger }`, constructed with an already-built `*feeds.Client` (not building `http.Client{}`/proxy pool inline), `Run(ctx)` calling `feeds.FetchIfDue(...)` then parsing the result and calling `r.repo.Upsert*`.
- Optional `internal/usecase/<source>_fetch.go` for source-specific fetch orchestration that's still allowed to call `feeds.*` helpers (see `coderules/internal/usecase/github_fetch.go` — it's *in* usecase but only calls `feeds.FetchIfDue`/`feeds.GitHubFetchRaw`, doesn't reimplement HTTP retry).

Concretely: for `ti` (L03), this means moving `Runner`/`NewRunner`/`Run`/`runKEV`/`runPTRSS`/etc. (currently in `internal/feeds/runner.go`) into a new `internal/usecase/runner.go`, and leaving behind in `internal/feeds/` only whatever raw per-feed HTTP fetch helpers those methods call into (check what's left after the move — likely small per-feed parse/fetch functions, not yet enumerated line-by-line here since the move itself will make the split obvious). For `lola`/`ds`/`vuln` (L04-L06), the fix is narrower: just relocate the inline `http.Client{}`/proxy-pool construction (already-cited line ranges) out of the usecase constructor into a `feeds.NewClient(...)`-style call, matching `coderules/internal/usecase/runner.go:35`'s constructor signature (`NewRunner(log, pub, opt, fc *feeds.Client, led *ledger.Store)` — the usecase receives the client, it doesn't build it).

---

## Priority backlog — suggested execution waves

Waves are ordered so each closes a coherent risk class before the next starts; independent waves may run in parallel.

### Wave A — P0 mechanical fixes (safe, isolated, no design decisions)
`G01`, `L01`, `L02`, `C01`, `C03`, `C04` — each is a single-file, single-concern change (add `recover()`, extract a usecase, add a port method, point a path at config, fix two doc snippets). No cross-team coordination needed.

### Wave B — P0 shutdown/gateway behavior (needs a real fix, not just markup)
`G02` → `G03` (fix in that order — G03 cannot be fixed correctly until G02 makes `Run()` cancellable), `C02` (aggregator degrade-gracefully policy — needs a decision on default behavior when engage is absent, see [C02](#c02)).

### Wave C — P1 doc cleanup (batchable, no code risk)
`C05`, `C06`, `C07` — same author pass, all about removing stale `engage/`-in-veil references post-split.

### Wave D — P1 discovery layering (templated fix, one PR pattern × 4)
`L03`, `L04`, `L05`, `L06`, `L08` — use `coderules`/`nuclei`/`sbom`'s existing correct `usecase` ↔ `internal/feeds` split as the template; `L08` (dedupe NVD structs onto `pkg/vuln/domain`) can land independently and first, since it's a pure DRY cleanup with a canonical type already available.

### Wave E — P2/P3 cleanup (low urgency, do opportunistically)
`G04`, `G05`, `G06`, `L07`, `O01`, `C08`, `C09`.

| Wave | IDs | Goal | Est. PRs |
|------|-----|------|----------|
| A | G01, L01, L02, C01, C03, C04 | Close P0 mechanical gaps | 5-6 |
| B | G02, G03, C02 | Fix shutdown-cleanup skip + gateway hard-fail | 2-3 |
| C | C05, C06, C07 | Remove stale post-split doc references | 1-2 |
| D | L03, L04, L05, L06, L08 | Unify usecase/feeds boundary across discovery sources | 4-5 |
| E | G04, G05, G06, L07, O01, C08, C09 | Opportunistic cleanup | 5-7 |

**Total estimated PRs:** 17-23.

---

## Code markup status

`TODO(veil-refactor <ID>)` / `FIXME(veil-refactor <ID>)` comments (FIXME = P0, TODO = P1 or lower) point back to this file. As of this audit:

**Marked in code:** all G/L/C findings addressed in refactor branches; grep `veil-refactor` should return 0 once fixes land.

Run `grep -rn "veil-refactor" --include="*.go" .` to list any remaining markers.
