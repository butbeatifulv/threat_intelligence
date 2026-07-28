# Coding style (Veil)

Conventions for the three runtime layers — **Discovery**, **pipeline**, **graph** — and shared contracts. When in doubt, mirror [discovery/harvest/internal/sources/ti/](../discovery/harvest/internal/sources/ti/) (Discovery) and [graph/ingest/internal/sources/ti/](../graph/ingest/internal/sources/ti/) (graph write path).

**Agents:** treat this file as source of truth; entry point [AGENTS.md](../AGENTS.md).

## Design principles

| Principle | Rule in this repo |
|-----------|-------------------|
| **CLEAN CODE** | Small functions, clear names, one level of abstraction; `cmd/` is wiring only |
| **DRY** | Shared fetch/ledger only in `discovery/`; normalize only in `pipeline/`; MERGE only in `knowledge/`; cross-layer types in `pkg/*` (wire envelopes, NATS helpers, TI helpers) |
| **KISS** | One long-running binary per layer at runtime; no speculative abstractions |
| **DDD** | Domain package per source module, no I/O (no Neo4j, NATS, HTTP clients in domain types) |

## Repository map

| Zone | Path | Role |
|------|------|------|
| **Knowledge** | [docs/](../) | Schemas, runtime, contracts — no application Go code |
| **Deploy** | [deploy/](../deploy/) | Compose per layer |
| **Discovery** | [discovery/](../discovery/) | Fetch + ledger → `scrape.>` |
| **Pipeline** | [pipeline/](../pipeline/) | `scrape.>` → NED → `ingest.>` |
| **Graph** | [knowledge/](../knowledge/) | Consume `ingest.>` → Neo4j; API/MCP read |
| **Veneno (pentest)** | [veneno](https://github.com/butbeautifulv/veneno) | Tool execution, workflows, reports (HTTP API/MCP) — separate repo |
| **Engage ingest bridge** | [pipeline/engage-events/](../pipeline/engage-events/) | NATS `engage.events.*` → `ingest.engage.*` (veneno → veil) |
| **Wire types** | [pkg/](../pkg/) | `harvest`, `commit`, `natsjet`, `ti/*`, `auth`, `engage/*` (events wire types for bridge) |

Layers **Discovery / pipeline / graph** communicate **only via NATS** and documented JSON schemas. **Veneno** calls graph only via **HTTP veil-api** (no Bolt, no NATS). No Go imports across `discovery/`, `pipeline/`, `knowledge/`. All layers may import `pkg/*`. NVD parse/map lives in [pipeline/pkg/nvd](../pipeline/pkg/nvd/) (pipeline only).

**Logical layers (v8 target)** — see [platform-architecture.md](../architecture/platform-architecture.md): **Discovery** (`discovery/`, P8h done), **Pipeline**, **Knowledge** (`knowledge/`, P8i), **Veneno** (pentest, external repo), shared **Report**, API/MCP façade. **Discovery `factory`** orchestrates sources; **veneno runner** executes catalog tools — share only **`pkg/exec`**, do not rename factory to runner. Split: [veil-veneno-split.md](../architecture/veil-veneno-split.md).

Layer-specific layout, env vars, and build commands:

| Layer | Docs |
|-------|------|
| Discovery | [discovery/README.md](../discovery/README.md), [discovery/harvest/README.md](../discovery/harvest/README.md) |
| Pipeline | [pipeline/README.md](../pipeline/README.md), [pipeline/ned/README.md](../pipeline/ned/README.md) |
| Graph | [knowledge/README.md](../knowledge/README.md), [knowledge/ingest/README.md](../knowledge/ingest/README.md) |

## Three runtime contexts

| Context | Code | NATS | Must not |
|---------|------|------|----------|
| **Discovery** | [discovery/](../discovery/) | Publish `scrape.>` | `commit`, Bolt, normalize |
| **Pipeline (NED)** | [pipeline/](../pipeline/) | `scrape.>` → `ingest.>` | HTTP feeds, Bolt, MERGE |
| **Graph** | [knowledge/](../knowledge/) | Consume `ingest.>` | `harvest`, feeds, Vitess |
| **Engage ingest** | [pipeline/engage-events/](../pipeline/engage-events/), [knowledge/ingest/internal/sources/engage/](../knowledge/ingest/internal/sources/engage/) | Consume veneno `engage.events.>` → `ingest.engage.*` → Neo4j | Bolt from bridge code, direct `scrape.>`, cross-layer Go imports |

Shared fetch policy (Discovery only): [discovery/harvest/internal/feeds](../discovery/harvest/internal/feeds/), [discovery/harvest/internal/ledger](../discovery/harvest/internal/ledger/) (`VITESS_DSN`, `SCRAPE_MIN_REFETCH_AFTER`, `SCRAPE_FORCE_REFETCH`).

Wire contracts: [ingest-contract.md](../contracts/ingest-contract.md). Go SOT: [pkg/harvest](../pkg/harvest/), [pkg/commit](../pkg/commit/).

---

## Layering (required packages)

Dependency direction — no cycles:

```
cmd/                    → wiring only (flags, env, construct usecase, Run)
domain/                 → entities, value objects, validation (no I/O)
internal/repository/    → ports (interfaces) — where used
internal/usecase/       → orchestration
internal/feeds/         → outbound HTTP/GitHub (scrape sources only)
storage/                → adapters at module root (Neo4j in graph; pub in layer connector)
```

`internal/*` is private to the Go module. Code another binary must import lives outside `internal/` (e.g. `storage/`, `scrapesource/`).

### Domain package paths (pkg SOT)

Shared entities live under **`pkg/*/domain/`** (single source of truth). Layers import `pkg` and keep only source-specific adapters (transform, storage, I/O). Full map: [domain-contour.md](../architecture/domain-contour.md).

| Package | Types | Layers |
|---------|-------|--------|
| `pkg/ti/domain` | IOC, Actor, Campaign, … | scrape TI, pipeline NED, graph ingest TI |
| `pkg/vuln/domain`, `pkg/lola/domain` | CVE, STIX artifacts | scrape, pipeline, graph |
| `pkg/{ds,sbom,nuclei,coderules}/domain` | AppSec entities | scrape, pipeline, graph |
| `pkg/engage/domain/{report,job,tool,target}` | Findings, jobs, tool spec, scan target | engage serve |
| `pkg/ti/{validate,ids,normalize}` | Normalization (pipeline only for normalize) | pipeline NED |

**Layer-local domain** (not in `pkg/`): knowledge serve read DTOs — see [domain-contour.md](domain-contour.md#knowledge-serve-read-dtos-layer-local-not-in-pkg) (`knowledge/connector/query`, `knowledge/serve/internal/usecase`).

**Rule:** new cross-layer entity → add to `pkg/<area>/domain` with `*_test.go`; do not duplicate `type IOC struct` in scrape/pipeline/graph.

Pipeline/scripts boundary: [scripts/README.md](../scripts/README.md) (`ops/`, `graph-pack/`, `test/`, `housekeeping/` — Neo4j housekeeping is not NED wire dedup). Graph packs: [docs/contracts/graph-pack.md](../contracts/graph-pack.md).

---

## PR checklist

Before merge, verify all items that apply to your layer:

| Rule | Scrape | Pipeline (NED) | Graph |
|------|--------|----------------|-------|
| `cmd/` has no business logic (no HTTP, Cypher, per-source transform) | ✓ | ✓ | ✓ |
| `usecase` has no NATS subject strings | ✓ | ✓ | ✓ |
| No `commit` / Bolt in scrape | ✓ | — | — |
| No `harvest` / feeds / Vitess in graph | — | — | ✓ |
| No cross-layer Go imports | ✓ | ✓ | ✓ |
| NVD parse only in pipeline (`pipeline/pkg/nvd`) | harvest publishes raw page only | ✓ enrich in `sources/vuln/enrich` | ingest does not re-parse NVD |
| Graph ingest does not import `pipeline/pkg/ti/normalize` | — | NED normalizes TI | ✓ |
| Idempotency keys via `pkg/commit` helpers only | — | ✓ | ✓ |
| `knowledge/serve` does not import NATS or scrape | — | — | ✓ |
| Shared entities in `pkg/*/domain` (no duplicate structs) | import `pkg` | import `pkg` | import `pkg` |
| `make test-platform-p7` when changing domain contour | ✓ | ✓ | ✓ |

### Agent / CI closure

For automated agents and maintainers, before marking work done:

| Step | Command / doc |
|------|----------------|
| Platform P7 (pkg domain + bus slices, no Docker) | `make test-platform-p7` — required when touching `pkg/*/domain`, `pkg/ti/*`, layer `domain/` or `usecase/` on the Veil contour |
| New or changed `pkg/*` logic | `make test-pkg-cover` — all pkg modules + [pkg-cover.sh](../../scripts/test/pkg-cover.sh) floors; add `*_test.go` in the same PR ([pkg-test-coverage.md](../development/pkg-test-coverage.md)) |
| Tests (touched layers) | `make test-discovery`, `make test-pipeline`, `make test-graph`; graph read/auth/MCP: `make test-graph-serve` |
| Graph read smoke (Docker) | `make test-graph-read-smoke` — no scrape/NATS |
| Graph version (ingest paths) | `./scripts/release/bump-graph-version.sh patch` → updates [versions.env](../versions.env) |
| Verify bump when required | `make check-graph-version` |
| Commit + push | See [AGENTS.md](../AGENTS.md) |

Ingest-affecting paths: `discovery/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `graph/ingest/internal/sources/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`.

---

## Markdown links (GitHub)

GitHub shows a folder icon only when the link target is a directory **and** the URL ends with `/`.

| Target | Link form |
|--------|-----------|
| Directory | `[label](../path/to/dir/)` — trailing `/` required |
| Layer README | `[Scrape](../discovery/README.md)` |
| File | `[compose.yml](../deploy/discovery/compose.yml)` — no trailing `/` |

Lint: `./scripts/housekeeping/lint-markdown-dir-links.sh`

---

## Go style (Google)

Follow the [Google Go Style Guide](https://google.github.io/styleguide/go/guide) and [best practices](https://google.github.io/styleguide/go/best-practices):

| Principle | In this repo |
|-----------|----------------|
| **Clarity** | Names and structure explain what and why; comments for rationale, not restating code |
| **Simplicity** | Prefer standard library over extra abstractions |
| **Consistency** | `gofmt`, **MixedCaps**, match neighboring files |
| **Package names** | Short, no `v1` in directory names; version only in `schema_version` |
| **Avoid repetition** | Do not repeat package name in exported symbols (`harvest.Envelope`, not `harvest.HarvestEnvelope`) |
| **Errors** | Wrap with `%w`; stable prefixes (`harvest:`, `commit:`); no `panic` in libraries; log once at `usecase` or `cmd` |
| **Tests** | Table-driven where useful; `testdata/` next to the package; Neo4j integration: build tag `integration` |

`pkg/` layout: one module ([pkg/go.mod](../pkg/go.mod)); scrape-only helpers under [discovery/pkg/](../discovery/pkg/).

---

## Naming

- Compose services / binaries: **`snake_case`** (`scrape_worker`, `pipeline_worker`, `ingest_worker`)
- NATS durable consumers: **`snake_case`** (`pipeline_worker`, `ingest_worker`)
- Go module path: `github.com/butbeautifulv/veil/...` (unchanged)

---

## Logging and lifecycle

- **`log/slog`** with structured attributes
- Long-running binaries: **`errgroup`** + cancel on **SIGINT/SIGTERM**
- Explicit **timeouts** on HTTP clients

---

## Configuration

- Environment variables with sensible defaults
- Document new env vars in [docs/architecture/threatintel-runtime.md](../architecture/threatintel-runtime.md)

---

## License

**MIT** — [LICENSE](../LICENSE), [CONTRIBUTING.md](../CONTRIBUTING.md)
