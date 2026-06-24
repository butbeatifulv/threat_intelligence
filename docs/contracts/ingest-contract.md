# Ingest contracts (harvest + commit + JetStream)

Veil uses **two NATS streams** between isolated layers:

```text
discovery/ → scrape.> (harvest) → pipeline/ → ingest.> (commit) → graph/ → Neo4j
```

**Go source of truth:** [pkg/harvest/](../pkg/harvest/), [pkg/commit/](../pkg/commit/). **JSON docs:** [schemas/harvest-envelope.json](schemas/harvest-envelope.json), [schemas/commit-envelope.json](schemas/commit-envelope.json) — update manually when pkg types change.

## harvest (Discovery → pipeline)

- **Go:** [pkg/harvest/](../pkg/harvest/)
- **Stream:** `SCRAPE`, subjects `scrape.>`
- **Publisher:** [discovery/harvest/](../discovery/harvest/) (`cmd/scrape_worker`)
- **Consumer:** [pipeline/ned/](../pipeline/ned/) (`cmd/pipeline_worker`)
- **Dedup (optional):** `Nats-Msg-Id` = `content_key`

## commit (pipeline → graph)

- **Go:** [pkg/commit/](../pkg/commit/) (pipeline + graph)
- **Stream:** `INGEST`, subjects `ingest.>`
- **Publisher:** pipeline via [pipeline/connector/](../pipeline/connector/)
- **Consumer:** [knowledge/ingest/](../knowledge/ingest/) (`cmd/ingest_worker`)
- **Dedup:** `Nats-Msg-Id` = `idempotency_key`

### TI commit payloads (NED → graph)

- Upsert kinds (`ti_ioc`, `ti_campaign`, …): payload is **already normalized by NED** (`pkg/ti/normalize`).
- Graph ingest **does not re-normalize**; Neo4j node `id` for IOC/Actor/Report comes from `idempotency_key` (`ti:ioc:…`, `ti:actor:…`, `ti:report:…`) via [pkg/commit/ti_node.go](../pkg/commit/ti_node.go).
- NVD CWE/CPE parsing runs only in pipeline ([pipeline/pkg/nvd/](../pipeline/pkg/nvd/)); harvest publishes raw `scrape_nvd_page` only.

### Engage commit payloads (cross-layer, optional)

**Roles after veil/veneno split:**

| Role | Repo | Component |
|------|------|-----------|
| **Publisher** | **veneno** | `engage/serve` → `engage.events.*` |
| **Bridge** | **veil** | `pipeline/engage-events` |
| **Consumer** | **veil** | `knowledge/ingest` source `engage` |

Veneno does **not** publish directly to `ingest.>`. When `ENGAGE_EVENTS_NATS_ENABLED=1`, veneno publishes JSON to stream **`ENGAGE_EVENTS`** (`engage.events.>`). [pipeline/engage-events](../pipeline/engage-events/) consumes and republishes `commit.Envelope` messages:

**Wire schemas:** [engage-events-audit.json](schemas/engage-events-audit.json), [engage-events-finding.json](schemas/engage-events-finding.json) → [ingest-engage-tool-run.json](schemas/ingest-engage-tool-run.json), [ingest-engage-finding.json](schemas/ingest-engage-finding.json).

| Engage subject | Ingest subject | Kind | Graph labels |
|----------------|----------------|------|--------------|
| `engage.events.audit` | `ingest.engage.tool_run` | `engage_tool_run` | `EngageToolRun`, `EngageTarget` |
| `engage.events.finding` | `ingest.engage.finding` | `engage_finding` | `EngageFinding`, `EngageTarget` |

- **Source:** `engage` ([pkg/commit/envelope.go](../pkg/commit/envelope.go))
- **Bridge:** [pipeline/connector/nats/engage_consumer.go](../pipeline/connector/nats/engage_consumer.go)
- **Graph ingest:** [knowledge/ingest/internal/sources/engage/](../knowledge/ingest/internal/sources/engage/)
- **Idempotency keys:** `engage:run:{tool}:{target}:{at}` and `engage:finding:{tool}:{target}:{title}`

## Vitess crawl ledger (scrape only)

Persistent on host at `var/veil/ledger/mysql/` (bind mount). HTTP bodies: `var/veil/blobs/`. See [graph-pack.md](graph-pack.md#persistent-crawl-state-varveil).

| Variable | Meaning |
|----------|---------|
| `VITESS_DSN` | MySQL-compatible ledger ([discovery/harvest/internal/ledger](../discovery/harvest/internal/ledger/)) |
| `SCRAPE_MIN_REFETCH_AFTER` | Default `24h` |
| `SCRAPE_FORCE_REFETCH` | `1` = ignore ledger |
| `SCRAPE_CACHE_DIR` | Disk cache root (`/data/cache` in compose → `var/veil/blobs` on host) |

If ledger says skip but cache file is missing, [FetchIfDue](../discovery/harvest/internal/feeds/fetch.go) refetches over HTTP instead of failing silently.

## Deploy

Per-layer Compose: [deploy/discovery](../deploy/discovery/compose.yml), [deploy/pipeline](../deploy/pipeline/compose.yml), [deploy/graph](../deploy/graph/compose.yml). Full stack: include all three or use [deploy/graph/compose.full.yml](../deploy/graph/compose.full.yml).
