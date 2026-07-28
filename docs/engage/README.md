# Engage / pentest documentation

Moved to **[veneno](https://github.com/butbeautifulv/veneno)** (`docs/engage/`).

Veil retains the **ingest bridge** only:

- `pipeline/engage-events/` — NATS `engage.events.*` → `ingest.engage.*`
- `knowledge/ingest/internal/sources/engage/` — Neo4j `EngageToolRun` / `EngageFinding`
- `GET /v1/categories/engage/context` — read API for veneno

See [veil-veneno-split.md](../architecture/veil-veneno-split.md).
