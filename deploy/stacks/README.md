# Compose stack presets (SSOT)

Named **stacks** are the single source of truth for which `deploy/profiles/*.env` files and `deploy/**/compose*.yml` overlays belong together. Scripts such as `compose-up-full.sh` and smoke tests implement these chains; the YAML here documents and stabilizes them without replacing raw `docker compose`.

## Stacks vs raw compose

| Approach | When to use |
|----------|-------------|
| **Stack preset** (`*.yml` in this directory) | Pick a documented scenario (smoke, full-loop, secure, pentest). Copy the `compose` list and `profiles` into your shell or automation. |
| **Raw compose** | One-off debugging, partial layer up (`deploy/knowledge/compose.yml` only), or experimenting with overlays not yet captured in a stack file. |

Stacks do **not** invoke Docker themselves. Translate a preset to a command from the repo root:

```bash
# Example: full-loop (after reading deploy/stacks/full-loop.yml)
set -a && source deploy/profiles/smoke-minimal.env && set +a
docker compose \
  -f deploy/discovery/compose.yml \
  -f deploy/pipeline/compose.yml \
  -f deploy/knowledge/compose.yml \
  -f deploy/knowledge/compose.neo4j-publish.yml \
  -f deploy/engage/compose.yml \
  -f deploy/engage/compose.veil-stack.yml \
  up -d --build engage-api engage-events-worker
```

Prefer existing ops scripts when listed under `scripts:` in a stack file (they set `COMPOSE_FILES`, scaling, and health checks).

## Preset index

| File | Purpose |
|------|---------|
| [minimal.yml](minimal.yml) | Scrape E2E smoke (`smoke-minimal` profile); base for closed-loop pilots |
| [full-loop.yml](full-loop.yml) | Data plane + engage on shared NATS/Neo4j (`compose.veil-stack.yml`) |
| [unified-edge.yml](unified-edge.yml) | Full-loop + platform `veil-edge` nginx (P12b; `/v1/`, `/api/`, `/mcp/*`) |
| [secure-graph.yml](secure-graph.yml) | Graph nginx TLS + auth (`secure-graph.env`) |
| [secure-engage.yml](secure-engage.yml) | Engage nginx TLS + hardening (`secure-engage.env`) |
| [secure.yml](secure.yml) | Combined reference for both secure layers |
| [secure-unified.yml](secure-unified.yml) | Graph + engage on one platform nginx (`deploy/platform/nginx`) |
| [pentest-prod.yml](pentest-prod.yml) | Local prod-like pentest target (hardened graph + pentest overlays) |

## Profiles

Env profiles live in [../profiles/](../profiles/). A stack may list zero or more profile files; `source` them before `docker compose` (or rely on a script that calls `source_profile` from `scripts/lib/common.sh`).

| Profile | Typical stack |
|---------|----------------|
| `smoke-minimal.env` | `minimal`, `full-loop` |
| `secure-graph.env` | `secure-graph`, `secure` |
| `secure-engage.env` | `secure-engage`, `secure` |
| `secure-graph.env` + `secure-engage.env` | `secure-unified` (both sourced) |
| `pentest-prod.env` | `pentest-prod` |

Graph-pack crawl profiles (`full-enrich`, `fast-rich`, `incremental-pack`, `full-rebuild`) are used by [scripts/graph-pack/](../../scripts/graph-pack/) and are not stack presets yet.

## YAML fields

| Field | Meaning |
|-------|---------|
| `name` | Short identifier |
| `description` | Human-readable intent |
| `profiles` | Paths to `deploy/profiles/*.env` (repo-relative) |
| `compose` | Ordered compose file chain (repo-relative) |
| `optional_compose` | Extra `-f` files under conditions (e.g. scrape partition) |
| `compose_profiles` | Docker Compose `--profile` values (e.g. `mcp`) |
| `services_up` | Suggested service subset when not using `compose-up-*.sh` |
| `scripts` | Repo-relative helpers that implement this stack |
| `notes` | Warnings, port defaults, mutual exclusions |
| `stacks` | (combined presets only) Nested knowledge/engage chains |

## Related docs

- [deploy/README.md](../README.md) — per-layer compose layout
- [docs/deploy/deploy-secure.md](../../docs/deploy/deploy-secure.md) — secure knowledge/engage runtime
- [docs/architecture/platform-unified-access.md](../../docs/architecture/platform-unified-access.md) — unified edge paths and RBAC
- [docs/engage/engage-runtime.md](../../docs/engage/engage-runtime.md) — engage overlays and ports
