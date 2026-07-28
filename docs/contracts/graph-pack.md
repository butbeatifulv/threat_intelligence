# Graph pack (export, release, import)

Versioned Neo4j snapshots as **`veil-graph-vX.Y.Z.zip`** + `manifest.json`. GitHub release tag: **`veil-graph-vX.Y.Z`** (same as ZIP basename without `.zip`).

Current default version: **[versions.env](../versions.env)** (`GRAPH_PACK_VERSION`).

## Artifact layout

| File | Role |
|------|------|
| `manifest.json` | `schema` (`veil.graph-pack/1`), `graphVersion`, `sha256`, `createdAt` |
| `graph.cypher` | APOC `apoc.export.cypher.all` dump |

Legacy packs (`threat-intel-graph-*.zip`, schema `threat-intelligence.graph-pack/1`) still import via [scripts/graph-pack/import.sh](../scripts/graph-pack/import.sh).

## Build (local)

```bash
./scripts/graph-pack/export-cypher.sh
./scripts/graph-pack/build.sh
```

Output: `var/veil/graph/releases/veil-graph-vX.Y.Z.zip` (version from `versions.env` or `GRAPH_PACK_VERSION` env; paths via [scripts/lib/common.sh](../scripts/lib/common.sh)).

## Persistent crawl state (`var/veil`)

| Path | Role |
|------|------|
| `var/veil/blobs/` | HTTP response cache (L1); survives `compose down -v` on Neo4j/NATS |
| `var/veil/ledger/mysql/` | MySQL datadir for `crawl_resource` (where we fetched, content SHA) |
| `var/veil/graph/` | Working `graph.cypher` + `releases/*.zip` |

`compose down -v` removes only **ephemeral** volumes (`neo4j_data`, `nats_data`). Ledger and blobs stay on the host.

**Migrate from legacy `data/`:**

```bash
mkdir -p var/veil/{blobs,ledger/mysql,graph/releases}
[ -d data/cache ] && rsync -a data/cache/ var/veil/blobs/
[ -d data/neo4j_user_export ] && rsync -a data/neo4j_user_export/ var/veil/graph/
```

## Build profiles

| Profile | Script | Neo4j seed | Crawl |
|---------|--------|------------|-------|
| **Incremental** (recommended) | [profile-incremental-pack.sh](../scripts/graph-pack/profile-incremental-pack.sh) | Import `BASE_GRAPH_PACK_VERSION` (default `v0.4.7`) | Delta only (`SCRAPE_FORCE_REFETCH=0`) |
| **Fast-rich** | [profile-fast-rich.sh](../scripts/graph-pack/profile-fast-rich.sh) | Empty (`GRAPH_PACK_SKIP=1`) | Uses ledger/cache; no full refetch by default |
| **Full rebuild** | `profile-fast-rich.sh --full` | Empty | Wipes `var/veil` ledger+blobs + `SCRAPE_FORCE_REFETCH=1` |

Incremental profile: [deploy/profiles/incremental-pack.env](../deploy/profiles/incremental-pack.env). Fast-rich limits: [deploy/profiles/fast-rich.env](../deploy/profiles/fast-rich.env).

Crawl observability: `./scripts/crawl/status.sh`, backup: `./scripts/crawl/ledger-dump.sh`.

## Publish (GitHub)

```bash
./scripts/release/publish-graph-pack.sh
# or after manual build:
./scripts/release/publish-graph-pack.sh --skip-build
```

After ingest-affecting code changes, bump first: `./scripts/release/bump-graph-version.sh patch`.

Sets `GRAPH_PACK_DEFAULT_URL` for [graph-bootstrap](../deploy/knowledge/docker/graph-bootstrap.sh) (`GRAPH_PACK_DEFAULT_VERSION` from `versions.env`).

## Import

- **Compose bootstrap:** `graph-bootstrap` service (default download or `GRAPH_PACK_SKIP=1`)
- **Host:** `USE_DOCKER_COMPOSE=1 ./scripts/graph-pack/import.sh <url-or-zip>`
- **Local ZIP test:** [docker-compose.testpack.yml](../docker-compose.testpack.yml)

Manifest schema: [graph-pack-manifest.schema.json](graph-pack-manifest.schema.json).

## Related

- [scripts/README.md](../scripts/README.md) — script index
- [deploy/README.md](../deploy/README.md) — scaling, smoke
- [threatintel-runtime.md](../architecture/threatintel-runtime.md) — compose services and env
