# Threat Intel runtime (Docker Compose)

Veil runs as **three isolated layers** under [`deploy/`](../deploy/): **scrape** (fetch + ledger + NATS publish), **pipeline** (normalize → `ingest.>`), **graph** (Neo4j ingest + API). Default graph stack: **Neo4j** → **graph-bootstrap** → **HTTP API**. Full pipeline: [`./scripts/ops/compose-up-full.sh`](../scripts/ops/compose-up-full.sh) or merged compose files (see [deploy/README.md](../deploy/README.md)). Worker scaling: `PIPELINE_WORKER_SCALE`, `INGEST_WORKER_SCALE` in [deploy/README.md](../deploy/README.md).

## Ports

| Service | Port | Notes |
|---------|------|--------|
| Neo4j Browser | `${NEO4J_HTTP_PORT:-7474}` (host) | Bolt `${NEO4J_BOLT_PORT:-7687}`; map with `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` if defaults are busy |
| HTTP API | 8090 | `API_PORT` to override published port (dev compose) |
| MCP Streamable HTTP | 8091 | `MCP_HTTP_PORT`; service profile **`mcp`**, `MCP_HTTP_ENABLED=1` |
| nginx (TLS) | 443 | **Secure overlay only** — [compose.secure.yml](../deploy/knowledge/compose.secure.yml); see [deploy-secure.md](../deploy/deploy-secure.md) |
| Proxybroker | 8099 | Full stack only; `PROXYBROKER_PORT` |
| NATS client | `${NATS_CLIENT_PORT:-4222}` | Full stack (`compose-up-full`); maps container `4222` |
| NATS monitoring | `${NATS_MONITOR_PORT:-8222}` | HTTP on container `8222`; **`nats`** healthcheck uses **`http://127.0.0.1:8222/healthz`** |
| Engage API / MCP | 8890 / 8892 | Fourth layer — details in [engage-runtime.md](engage-runtime.md) |

## Build and environment checklist (repeatable builds)

Use this before a full scrape run or a reproducible graph pack.

1. **Full stack build** — `docker compose -f deploy/discovery/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml build` (or `./scripts/ops/compose-up-full.sh`). If builds fail with **EOF** or truncated module downloads, set **`GOPROXY`** for Docker builds or use a stable proxy.
2. **Bolt / Browser ports** — if `7687` or `7474` are taken, set **`NEO4J_BOLT_PORT`** / **`NEO4J_HTTP_PORT`** and stop duplicate Neo4j containers.
3. **Graph-only** — `docker compose up --build -d` then **`curl -sS http://localhost:${API_PORT:-8090}/health`**.
4. **Full pipeline** — `./scripts/ops/compose-up-full.sh`; **`scrape_worker`** waits for **`nats` healthy**; **`ingest_worker`** waits for **Neo4j** and **NATS**. TLS/EOF during `go mod download` in Docker is usually network/proxy, not application code.

## Compose service reference

Layer compose files: [deploy/discovery/compose.yml](../deploy/discovery/compose.yml), [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml), [deploy/graph/compose.yml](../deploy/graph/compose.yml). Root [docker-compose.yml](../docker-compose.yml) includes graph only. Optional local graph pack: [docker-compose.testpack.yml](../docker-compose.testpack.yml).

### neo4j

| | |
|--|--|
| **Profile** | *(none — starts with default `docker compose up`)* |
| **Image** | `neo4j:5` |
| **Purpose** | Graph database; APOC enabled for exports |
| **Volumes** | `neo4j_data` (ephemeral); host `var/veil/graph` → Neo4j import/export path |
| **Health** | `cypher-shell` `RETURN 1` |
| **Env** | `NEO4J_AUTH`, `NEO4J_PLUGINS` (APOC), `NEO4J_apoc_export_file_enabled=true` for export |

### graph-bootstrap

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [deploy/knowledge/docker/graph-bootstrap.Dockerfile](../deploy/knowledge/docker/graph-bootstrap.Dockerfile) |
| **Purpose** | One-shot import of a graph pack ZIP before API starts |
| **Depends on** | `neo4j` healthy |
| **Restart** | `no` |
| **Env** | `GRAPH_PACK_*`, `NEO4J_*` — see [Graph bootstrap (usage mode)](#graph-bootstrap-usage-mode) |

### api

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [deploy/knowledge/docker/api.Dockerfile](../deploy/knowledge/docker/api.Dockerfile) — distroless, non-root |
| **Purpose** | Categorical REST API over Neo4j |
| **Ports** | `${API_PORT:-8090}:8090` (dev); **none** on host in secure overlay |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Health** | `CMD ["/api", "healthcheck"]` (Neo4j ping); HTTP `GET /health` |
| **Env** | `API_LISTEN` (default `:8090`), `API_ENV`, `NEO4J_*`, `VEIL_REQUIRE_AUTH`, `CORS_ALLOWED_ORIGINS`, optional `AUTH_*` / `KEYCLOAK_*` — [auth-keycloak.md](auth-keycloak.md) |

### mcp

| | |
|--|--|
| **Profile** | `mcp` — opt-in (`docker compose --profile mcp up`) |
| **Build** | [deploy/knowledge/docker/mcp.Dockerfile](../deploy/knowledge/docker/mcp.Dockerfile) — distroless |
| **Purpose** | Stdio MCP + optional Streamable HTTP — [knowledge/serve/](../knowledge/serve/) |
| **Ports** | `${MCP_HTTP_PORT:-8091}` when `MCP_HTTP_ENABLED=1`; **none** on host in secure overlay |
| **Run** | **Host:** `cd knowledge/serve && go run ./cmd/mcp`. **Docker:** profile `mcp` or `docker run -i` — [mcp-agents.md](../agents/mcp-agents.md) |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Health** | `CMD ["/mcp", "healthcheck"]` |
| **Env** | `NEO4J_*`, `MCP_HTTP_*`, `MCP_HTTP_AUTH_STRICT`, `MCP_ENV`, optional `AUTH_*` / `MCP_ACCESS_TOKEN` — [auth-keycloak.md](auth-keycloak.md) |

### nginx (secure overlay)

| | |
|--|--|
| **Compose** | [deploy/knowledge/compose.secure.yml](../deploy/knowledge/compose.secure.yml) |
| **Build** | [deploy/knowledge/docker/nginx.Dockerfile](../deploy/knowledge/docker/nginx.Dockerfile) |
| **Purpose** | TLS termination; proxy `/` → API, `/mcp` → MCP; Neo4j not exposed |
| **Ports** | `${NGINX_HTTPS_PORT:-443}:443` |
| **TLS** | Mount `tls.crt` / `tls.key` — see [deploy/knowledge/nginx/certs/README.md](../deploy/knowledge/nginx/certs/README.md) |
| **Profile** | Use with [deploy/profiles/secure-graph.env](../deploy/profiles/secure-graph.env) — [deploy-secure.md](../deploy/deploy-secure.md) |

### nats (JetStream broker)

| | |
|--|--|
| **Compose** | [deploy/discovery/compose.yml](../deploy/discovery/compose.yml) (full stack) |
| **Image** | `nats:2.10-alpine` |
| **Command** | `-js -m 8222` (JetStream + monitoring) |
| **Purpose** | Two-stream bus: scrapers → **`scrape.>`** (`SCRAPE`); **pipeline_worker** → **`ingest.>`** (`INGEST`) |
| **Ports** | Client and monitoring (see [Ports](#ports)) |
| **Health** | `wget` on `http://127.0.0.1:8222/healthz` (JetStream monitoring) |
| **Streams** | **`SCRAPE`** (`scrape.>`), **`INGEST`** (`ingest.>`) — ensured by **scrapepub** / **pipeline_worker** / **ingest_worker** |

### pipeline_worker

| | |
|--|--|
| **Compose** | [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml) |
| **Build** | [deploy/pipeline/docker/pipeline_worker.Dockerfile](../deploy/pipeline/docker/pipeline_worker.Dockerfile) |
| **Module** | [pipeline/ned/README.md](../pipeline/ned/README.md) |
| **Purpose** | Pull **`scrape.>`**, normalize/dedup **`harvest`** → **`commit`**, publish **`ingest.>`** (per-domain subjects via `*_INGEST_SUBJECT`) |
| **Depends on** | **`nats` healthy** (when scrape layer is merged) |
| **Scale** | `PIPELINE_WORKER_SCALE` — shared durable `pipeline_worker` |
| **Env** | `NATS_URL`, `NATS_SCRAPE_STREAM`, `NATS_SCRAPE_DURABLE`, `NATS_SCRAPE_SUBSCRIBE_SUBJECT`, `PIPELINE_BATCH`, `PIPELINE_MAX_WAIT`, `DS_INGEST_SUBJECT`, `TI_INGEST_SUBJECT`, … |

### ingest_worker

| | |
|--|--|
| **Compose** | [deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml) |
| **Build** | [deploy/knowledge/docker/ingest_worker.Dockerfile](../deploy/knowledge/docker/ingest_worker.Dockerfile) |
| **Module** | [knowledge/ingest/README.md](../knowledge/ingest/README.md) |
| **Purpose** | Long-running **JetStream pull consumer**: reads `commit` from **`ingest.>`**, writes **Neo4j** (AppSec via `knowledge/ingest/internal/appsec/*`; ti/vuln/lola/ds via `knowledge/ingest/internal/sources/*`) |
| **Depends on** | `neo4j` healthy, **`nats` healthy** (full stack) |
| **Restart** | `unless-stopped` |
| **Scale** | `INGEST_WORKER_SCALE` — shared durable `ingest_worker` |
| **Env** | `NEO4J_*`, `NATS_URL`, `NATS_INGEST_STREAM`, `NATS_DURABLE`, `NATS_SUBSCRIBE_SUBJECT`, `INGEST_BATCH`, `INGEST_MAX_WAIT` |

Use **`ingest_worker`** whenever scrape producers run; without it, messages stay in JetStream until drained.

### proxybroker

| | |
|--|--|
| **Compose** | [deploy/discovery/compose.yml](../deploy/discovery/compose.yml) |
| **Build** | [deploy/discovery/docker/proxybroker.Dockerfile](../deploy/discovery/docker/proxybroker.Dockerfile) |
| **Purpose** | HTTP proxy pool for scrapers (`*_PROXY_URLS`) |
| **Ports** | `${PROXYBROKER_PORT:-8099}:8099` |

### scrape_worker (factory orchestrator)

| | |
|--|--|
| **Compose** | [deploy/discovery/compose.yml](../deploy/discovery/compose.yml) |
| **Build** | [deploy/discovery/docker/scrape_worker.Dockerfile](../deploy/discovery/docker/scrape_worker.Dockerfile) |
| **Module** | [discovery/harvest/README.md](../discovery/harvest/README.md) |
| **Purpose** | Runs selected sources via [discovery/harvest/internal/factory](../discovery/harvest/internal/factory/); publishes **`harvest`** to **`scrape.>`** (batch job, exits 0) |
| **Depends on** | **`nats` healthy** |
| **Partition** | `SCRAPE_WORKER_PARTITION=1` → `scrape_worker_fast` + `scrape_worker_slow` ([deploy/compose.scale.yml](../deploy/compose.scale.yml)) |
| **Env** | **`SCRAPE_SOURCES`**, **`TI_FEEDS`**, **`TI_JSONL_FILE`**, **`SBOM_CVE_LIST_FILE`**, **`VITESS_DSN`**, per-source scrape subjects; optional `GITHUB_TOKEN` (rate limits only) |

Sources live under [discovery/harvest/internal/sources/](../discovery/harvest/internal/sources/). They publish **`harvest`** only (no Bolt). **`pipeline_worker`** → **`commit`**; **`ingest_worker`** → Neo4j. See [discovery/README.md](../discovery/README.md).

### NATS subjects

| Variable | Default | Meaning |
|----------|---------|--------|
| `SCRAPE_SOURCES` | `ds,vuln,lola,ti,sbom,coderules,nuclei` | Comma-separated sources for **`scrape_worker`** |
| `TI_FEEDS` | `kev,urlhaus,threatfox,malwarebazaar,feodo` | Feed list when `ti` is enabled |
| `TI_JSONL_FILE` | `/app/example.jsonl` (Compose) | Optional JSONL path; empty to skip |
| `NATS_URL` | `nats://nats:4222` in Compose | NATS client URL |
| `VULN_SCRAPE_SUBJECT` | `scrape.vuln.events` | Scraper publish subject for **`vuln`** |
| `LOLA_SCRAPE_SUBJECT` | `scrape.lola.events` | Scraper publish for **`lola`** |
| `DS_SCRAPE_SUBJECT` | `scrape.ds.events` | Scraper publish for **`ds`** |
| `TI_SCRAPE_SUBJECT` | `scrape.ti.events` | Scraper publish for **`ti`** |
| `SBOM_SCRAPE_SUBJECT` | `scrape.appsec.sbom` | Scraper publish for `sbom` |
| `CODERULES_SCRAPE_SUBJECT` | `scrape.appsec.coderules` | Scraper publish for `coderules` |
| `NUCLEI_SCRAPE_SUBJECT` | `scrape.appsec.nuclei` | Scraper publish for `nuclei` |
| `DS_INGEST_SUBJECT` | `ingest.ds.events` | **pipeline_worker** publish for DS |
| `TI_INGEST_SUBJECT` | `ingest.ti.events` | pipeline publish for TI |
| `VULN_INGEST_SUBJECT` | `ingest.vuln.events` | pipeline publish for vuln |
| `LOLA_INGEST_SUBJECT` | `ingest.lola.events` | pipeline publish for lola |
| `SBOM_INGEST_SUBJECT` | `ingest.appsec.sbom` | pipeline publish for sbom |
| `CODERULES_INGEST_SUBJECT` | `ingest.appsec.coderules` | pipeline publish for coderules |
| `NUCLEI_INGEST_SUBJECT` | `ingest.appsec.nuclei` | pipeline publish for nuclei |
| `SBOM_CVE_LIST_FILE` | `/fixtures/cve_list_seed.txt` in Compose | CVE list for OSV (one `CVE-…` per line; `#` comments allowed) |
| `SBOM_CVE_LIST_URL` | empty | Alternative CVE list URL if file unset |
| `VITESS_DSN` / `MYSQL_DSN` | `veil:veilpass@tcp(crawl-db:3306)/veil_ledger` in scrape compose | Crawl ledger ([discovery/harvest/internal/ledger](../discovery/harvest/internal/ledger/)); records URL + `content_sha256` |
| `SCRAPE_MIN_REFETCH_AFTER` | `24h` | Min refetch interval (`periodic` policy) |
| `SCRAPE_FORCE_REFETCH` | `0` | `1` = ignore ledger (full refetch). Graph-pack profiles use `0` except `full-rebuild` / `--full`. |
| `SCRAPE_CACHE_DIR` | `/data/cache` in compose | Host bind: `var/veil/blobs` |
| `LOFTS_SKIP_ON_ERROR` | unset | `true` = LOFTS fetch errors are warnings (do not fail `lola`) |
| `SCRAPE_FAIL_FAST` | unset | `1` = stop all sources on first source error |
| `NATS_INGEST_STREAM` | `INGEST` | Stream name (worker) |
| `NATS_DURABLE` | `ingest_worker` | Durable consumer name |
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.>` | Worker pull filter (AppSec, TI, vuln, lola, ds, …) |
| `INGEST_BATCH` | `10` | Max messages per fetch |
| `INGEST_MAX_WAIT` | `5s` | Fetch wait |

JetStream dedup: **`Nats-Msg-Id`** from envelope **`idempotency_key`** ([pipeline/connector/](../pipeline/connector/), [pkg/commit/](../pkg/commit/)).

Contract details: [docs/contracts/ingest-contract.md](../contracts/ingest-contract.md).

## Graph bootstrap (usage mode)

Init container `graph-bootstrap` runs once after Neo4j is healthy. Pack resolution order:

1. **`/pack/releases/veil-graph-${GRAPH_PACK_DEFAULT_VERSION}.zip`** (host `var/veil/graph/releases/`, bind-mounted).
2. Bind mount **`/pack/host.zip`** (e.g. [docker-compose.testpack.yml](../docker-compose.testpack.yml)).
3. **`GRAPH_PACK_FILE`** — path *inside the container* to a `.zip`.
4. **`GRAPH_PACK_URL`** — HTTP(S) URL to a pack ZIP.
5. If **`GRAPH_PACK_DEFAULT=1`**: download **`GRAPH_PACK_DEFAULT_URL`** or the built-in release asset.

Skip import entirely: **`GRAPH_PACK_SKIP=1`**.

| Variable | Default | Meaning |
|----------|---------|--------|
| `GRAPH_PACK_SKIP` | `0` | `1` = exit 0 without importing |
| `GRAPH_PACK_DEFAULT` | `1` | `0` = do not download the default release ZIP when no file/URL |
| `GRAPH_PACK_URL` | empty | HTTP(S) URL of the pack ZIP |
| `GRAPH_PACK_DEFAULT_URL` | `veil-graph-v0.4.7` asset | Overrides the default download URL when `GRAPH_PACK_DEFAULT=1` |
| `GRAPH_PACK_DEFAULT_VERSION` | `v0.4.7` | Used to build default URL when `GRAPH_PACK_DEFAULT_URL` unset |
| `GRAPH_PACK_FILE` | empty | Path **inside** the bootstrap container (mount a volume if needed) |

Compose passes these from the host for `graph-bootstrap` (see [docker-compose.yml](../docker-compose.yml) `environment`).

Checksum: `manifest.json` `sha256` must match `graph.cypher` (same rules as [scripts/graph-pack/import.sh](../scripts/graph-pack/import.sh)).

## Full scrape stack

```bash
./scripts/ops/compose-up-full.sh
```

Worker scaling and scrape partition: [deploy/README.md](../deploy/README.md#worker-scaling-parallel-nats-consumers). Graph pack export, fast-rich profile (~25 min), and releases: [docs/contracts/graph-pack.md](../contracts/graph-pack.md).

Neo4j export requires `NEO4J_apoc_export_file_enabled=true` ([deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml)).

### Smoke checklist

1. **Default stack (no scrape):** `docker compose up --build -d` → wait for **`api` healthy** → `curl` **`/health`** and a few **`/v1/...`** calls (see [curl examples](#curl-examples)).
2. **Platform bus (unit):** `make test-platform-p0` — `pkg/natsjet`, pipeline/graph ingest consumer tests (no Docker).
3. **Engage closed loop (Docker):** `make test-platform-closed-loop` — veil stack + engage overlay, act→ingest→`target-graph` ([platform-closed-loop-pilot.md](platform-closed-loop-pilot.md)).
4. **Graph read path (no scrape/NATS/ingest):** `make test-graph-read-smoke` or `./scripts/test/smoke-graph-read.sh --up` — overlay [compose.graph-read.yml](../deploy/graph/compose.graph-read.yml), `GRAPH_PACK_SKIP=1`, optional MCP HTTP.
5. **Graph pack without GitHub download:** place **`veil-graph-v0.4.5.zip`** under `var/veil/graph/releases/` and run `docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d` (see [docker-compose.testpack.yml](../docker-compose.testpack.yml)).
6. **Scrape + NATS:** `./scripts/test/smoke-discovery-e2e.sh --up`; confirm JetStream drains and Neo4j gains nodes (see [discovery/README.md](../discovery/README.md), [graph/README.md](../graph/README.md)).
7. **Release asset:** the default URL in [deploy/graph/docker/graph-bootstrap.sh](../deploy/graph/docker/graph-bootstrap.sh) must point at a ZIP that contains **`manifest.json`** + **`graph.cypher`** with matching **`sha256`**. Bump version and URLs if the dump changes.
8. **Secure prod overlay:** TLS certs + Keycloak env → [deploy-secure.md](../deploy/deploy-secure.md).

### E2E scrape smoke (slice 8 v2)

Automated checks for the full scrape → pipeline → ingest path:

```bash
./scripts/test/smoke-discovery-e2e.sh --up          # compose up scrape stack, wait for scrape_worker exit
./scripts/test/smoke-discovery-e2e.sh               # NATS health, crawl_resource rows, Neo4j counts, API /health
./scripts/test/smoke-discovery-e2e.sh --restart-scrape   # pass 2: ledger unchanged / skip publish
```

Env overrides: `SCRAPE_SVC`, `PIPELINE_SVC`, `INGEST_SVC`, `PIPELINE_WORKER_SCALE`, `INGEST_WORKER_SCALE`, `SCRAPE_WORKER_PARTITION`, `NATS_MON`, `API_URL`, `CRAWL_MYSQL`, `SMOKE_WAIT_SEC`.

## HTTP API (categorical)

Base URL: `http://localhost:8090` (with default compose).

- `GET /health`
- `GET /v1/categories` — product categories (`vuln`, `ti`, `detection`, `lola`, `mitre`, `sbom`, `code_rules`, `dast`, `engage`) and Neo4j label sets. **`engage`** holds `EngageToolRun` / `EngageFinding` from the engage events bus.
- `GET /v1/categories/{category}/kinds` — labels present in the graph with counts.
- `GET /v1/categories/{category}/nodes?kind=Vulnerability&limit=50&offset=0`
- `GET /v1/categories/{category}/search?q=cve&kind=&limit=50`
- `GET /v1/nodes/{id}` — elementId or `id` / `cve` / `uri` / `link`
- `GET /v1/nodes/{id}/neighbors?depth=1&limit=500`
- `GET /v1/kinds` — all distinct labels (legacy discovery)

OpenAPI sketch: [openapi.yaml](openapi.yaml).

### curl examples

Assuming default `API_PORT=8090` after `docker compose up`:

```bash
curl -sS http://localhost:8090/health
curl -sS http://localhost:8090/v1/categories | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/kinds' | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/nodes?kind=Vulnerability&limit=5' | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/search?q=cve&limit=5' | jq .
curl -sS 'http://localhost:8090/v1/categories/engage/search?q=example.com&limit=10' | jq .
curl -sS 'http://localhost:8090/v1/kinds' | jq .
```

Replace `vuln` / `Vulnerability` with other [categories](../knowledge/connector/query/categories.go) and labels as needed.

## MCP (stdio + Streamable HTTP)

Same categorical logic as the API. Server name **veil-mcp**. Full agent setup: [mcp-agents.md](../agents/mcp-agents.md).

From source (host, against compose Neo4j on localhost:7687):

```bash
cd knowledge/serve && go run ./cmd/mcp
# optional HTTP on :8091:
MCP_HTTP_ENABLED=1 go run ./cmd/mcp
```

HTTP endpoint: `POST http://localhost:8091/mcp` (protocol `2025-03-26`, JSON or SSE).

Smoke: `./scripts/smoke/mcp-smoke.sh` (requires running Neo4j).

Category-first tools: `ti_list_categories`, `ti_list_kinds_in_category`, `ti_nodes_by_category`, `ti_search_in_category`, `ti_get_node`, `ti_neighbors`, `ti_health`. Legacy tools (`ti_list_kinds`, `ti_get_nodes_by_kind`, `ti_search`) are deprecated.

## Optional: local pack file

**Quick smoke test with the repo’s sample pack** (no download):

```bash
docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d
```

See [docker-compose.testpack.yml](../docker-compose.testpack.yml) (bind-mounts `var/veil/graph/releases/veil-graph-v0.4.7.zip` as `/pack/host.zip` and sets `GRAPH_PACK_DEFAULT=0`).

Re-importing the same pack into **non-empty** Neo4j (existing constraints) will fail. Use `./scripts/ops/compose-down-ephemeral.sh` (keeps `var/veil` ledger/blobs) before a clean import. **`ingest_worker`** waits for **`graph-bootstrap`** to finish ([deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml)).

Create `docker-compose.override.yml` (gitignored by convention or not committed) with:

```yaml
services:
  graph-bootstrap:
    volumes:
      - ./my-pack.zip:/pack/host.zip:ro
```

The bootstrap script copies `/pack/host.zip` when it exists and is non-empty.

## Observability

Veil exports Prometheus RED metrics on graph HTTP services, OpenTelemetry traces to Tempo, and JSON logs with `trace_id` for Loki. See [OBSERVABILITY.md](../OBSERVABILITY.md) for env vars, k3s deploy, Grafana dashboards, and egregore cross-service trace verification.
