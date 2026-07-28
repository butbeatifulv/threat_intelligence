# Pipeline layer (NED)

Part of the four-layer Veil stack — [README.md](../README.md#architecture). Architecture rules: [docs/agents/coding-style.md](../docs/agents/coding-style.md).

Consumes `scrape.>`, applies **NED** (normalize, enrichment, deduplication), publishes `ingest.>`.

| Module | Path | Role |
|--------|------|------|
| **connector** | [connector/](connector/) | NATS JetStream publish + stream ensure |
| **ned** | [ned/](ned/) | Transform worker (`pipeline_worker`) |
| **engage-events** | [engage-events/](engage-events/) | Bridge `engage.events.>` → `ingest.engage.tool_run` / `ingest.engage.finding` (Phase 12–13) |

- **Wire types:** [pkg/harvest/](../pkg/harvest/), [pkg/commit/](../pkg/commit/)
- **Build:** `make test-pipeline` or:

```bash
cd pipeline/ned && go build -o bin/pipeline_worker ./cmd/pipeline_worker
```

- **Deploy:** [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml)

## ned layout

```
ned/
  cmd/pipeline_worker/     # thin main
  internal/
    components/            # NATS DI
    consumer/              # scrape pull loop
    transform/             # route by source
    dedup/                 # publish with idempotency keys
    sources/{ti,vuln,lola,ds,appsec}/
```
