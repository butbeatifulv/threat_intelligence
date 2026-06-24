# Split R0 inventory — veil → veneno

Generated for [veil-veneno-split ADR](../architecture/veil-veneno-split.md).

## Non-engage importers of `pkg/engage` / `pkg/exec` / `pkg/decision` / `pkg/report`

| Path | Imports | Action after split |
|------|---------|-------------------|
| `pipeline/connector/nats/engage_consumer.go` | `pkg/engage/events` | **Keep** `pkg/engage/events` in veil |
| `knowledge/connector/query/engage_context.go` | `pkg/engage/hostnorm` | **Keep** `pkg/engage/hostnorm` in veil |
| `pkg/domain/engage.go` | `pkg/engage/domain/*` aliases | Remove aliases or link to contract-only types |
| `pkg/report/*` | `pkg/engage/domain/report` | **Move** to veneno with `pkg/report` |
| `pkg/go.mod` | `pkg/engage` (report PDF) | Trim veil `pkg/` after report moves |

All other importers are under `engage/` or `pkg/engage|exec|decision|report` → move to veneno.

## File manifest (engage → veneno)

| Source (veil) | Destination (veneno) | ~files |
|---------------|------------------------|--------|
| `engage/` | `engage/` | 120+ |
| `deploy/engage/` | `deploy/engage/` | 80+ |
| `docs/engage/` | `docs/engage/` | 15+ |
| `scripts/engage/` | `scripts/engage/` | 10+ |
| `scripts/eval/pentest-veil-mcp.sh` | `scripts/eval/pentest-veneno-mcp.sh` | 1 |
| `pkg/engage/` (except events+hostnorm in veil) | `pkg/engage/` | 15 |
| `pkg/exec/` | `pkg/exec/` | all |
| `pkg/decision/` | `pkg/decision/` | all |
| `pkg/report/` | `pkg/report/` | all |
| `pkg/auth/`, `pkg/api/`, `pkg/mcp/` | same (engage slices) | subset |
| `deploy/helm/veil/templates/engage-*` | `deploy/helm/veneno/templates/` | 4 |

**Stays in veil:** `pipeline/engage-events/`, `knowledge/ingest/.../engage/`, `pkg/engage/events/`, `pkg/engage/hostnorm/`, `pkg/commit` engage payloads.

## Sync list (manual until cxado-libs)

Duplicate in veneno, do not delete from veil until bridge tested:

- `pkg/engage/events` — veneno publisher; veil consumer (bridge)
- `pkg/auth`, `pkg/api`, `pkg/mcp` — transport shared pattern
