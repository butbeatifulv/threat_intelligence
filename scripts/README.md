# Scripts

Host-side helpers for **Neo4j export**, **graph pack** releases, **stack ops**, **E2E smoke**, and **housekeeping**. Not the pipeline NED runtime ([pipeline/ned](../pipeline/ned/)).

Shared library: [lib/common.sh](lib/common.sh) (`COMPOSE_FILES`, `compose()`, pack naming `veil-graph-v*`).

## Layout

| Path | Purpose |
|------|---------|
| [lib/common.sh](lib/common.sh) | Compose helpers, `veil-graph-v*` pack names, profiles |
| [test/lib/smoke.sh](test/lib/smoke.sh) | Shared smoke helpers (`smoke_skip_no_docker`, `smoke_wait_http`) |
| [test/lib/unit.sh](test/lib/unit.sh) | Shell test helpers (`unit_skip_no_go`, JSON field asserts) |
| [ops/compose-up-full.sh](ops/compose-up-full.sh) | Full stack up (optional worker scale / scrape partition) |
| [ops/compose-down-ephemeral.sh](ops/compose-down-ephemeral.sh) | `down` keeping `var/veil` ledger + blobs |
| [crawl/status.sh](crawl/status.sh) | Ledger summary + blob dir size |
| [crawl/ledger-dump.sh](crawl/ledger-dump.sh) | Export `crawl_resource` to JSON |
| [graph-pack/export-cypher.sh](graph-pack/export-cypher.sh) | APOC export → `var/veil/graph/graph.cypher` |
| [graph-pack/build.sh](graph-pack/build.sh) | ZIP + `manifest.json` → `var/veil/graph/releases/veil-graph-vX.zip` |
| [graph-pack/import.sh](graph-pack/import.sh) | Import pack (supports legacy `threat-intel-graph-*.zip`) |
| [graph-pack/profile-incremental-pack.sh](graph-pack/profile-incremental-pack.sh) | Seed from BASE pack + delta crawl (recommended) |
| [graph-pack/profile-fast-rich.sh](graph-pack/profile-fast-rich.sh) | ~25 min crawl; `--full` wipes `var/veil` + force refetch |
| [release/publish-graph-pack.sh](release/publish-graph-pack.sh) | Build + `gh release create veil-graph-vX` |
| [release/bump-graph-version.sh](release/bump-graph-version.sh) | Bump `GRAPH_PACK_VERSION` in [versions.env](../versions.env) |
| [release/check-graph-version-bump.sh](release/check-graph-version-bump.sh) | Fail if ingest paths changed without version bump |
| [housekeeping/sync-github-metadata.sh](housekeeping/sync-github-metadata.sh) | Push [.github/repo-description.txt](../.github/repo-description.txt) to GitHub |
| [housekeeping/lint-markdown-dir-links.sh](housekeeping/lint-markdown-dir-links.sh) | Lint directory links (trailing `/`) in `*.md` |
| [test/smoke-discovery-e2e.sh](test/smoke-discovery-e2e.sh) | E2E smoke (default profile [deploy/profiles/smoke-minimal.env](../deploy/profiles/smoke-minimal.env)) |
| [housekeeping/audit-repo-refs.sh](housekeeping/audit-repo-refs.sh) | Classify scripts as MAKE/CI/KEEP/ORPHAN (heuristic) |
| [test/smoke-graph-read.sh](test/smoke-graph-read.sh) | Graph read smoke: Neo4j + API + MCP HTTP (no scrape/NATS) |
| [test/smoke-unified-edge.sh](test/smoke-unified-edge.sh) | P12 unified TLS nginx edge (veneno upstream for `/api`, `/mcp/engage`) |
| [mcp/run-veil-mcp.sh](mcp/run-veil-mcp.sh) | MCP stdio launcher for agents (logs on stderr) |
| [smoke/mcp-smoke.sh](smoke/mcp-smoke.sh) | MCP stdio smoke against local Neo4j |
| [test/smoke-engage-events-pipeline.sh](test/smoke-engage-events-pipeline.sh) | Veneno `engage.events` → veil ingest bridge |
| [test/verify-nvd-enrichment.sh](test/verify-nvd-enrichment.sh) | Cypher QA for NVD CWE/CPE |
| [housekeeping/graph-dedup-cleanup.sh](housekeeping/graph-dedup-cleanup.sh) | Post-ingest Neo4j dedup |

Platform Docker smokes source [test/lib/smoke.sh](test/lib/smoke.sh) for docker skip and HTTP health polls:

```bash
source "$(dirname "$0")/lib/smoke.sh"
smoke_skip_no_docker
smoke_wait_http "${API_URL}/health" 120 "veil-api" 2
```

Deploy profiles: [deploy/profiles/](../deploy/profiles/) (`smoke-minimal`, `secure-graph`). Runtime: [docs/architecture/threatintel-runtime.md](../docs/architecture/threatintel-runtime.md). Secure deploy: [docs/deploy/deploy-secure.md](../docs/deploy/deploy-secure.md). Graph pack workflow: [docs/contracts/graph-pack.md](../docs/contracts/graph-pack.md).

## Graph pack workflow

```bash
./scripts/graph-pack/profile-incremental-pack.sh   # seed + delta crawl (recommended)
# or: ./scripts/graph-pack/profile-fast-rich.sh
./scripts/crawl/status.sh
./scripts/graph-pack/export-cypher.sh
./scripts/graph-pack/build.sh
./scripts/release/publish-graph-pack.sh --skip-build
```

Import locally:

```bash
USE_DOCKER_COMPOSE=1 ./scripts/graph-pack/import.sh \
  var/veil/graph/releases/veil-graph-v0.4.7.zip
```

## Quick commands

```bash
./scripts/ops/compose-up-full.sh
./scripts/test/smoke-discovery-e2e.sh --up
make test-graph-read-smoke
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/ops/compose-up-full.sh
```
