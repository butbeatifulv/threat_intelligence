# Neo4j Enterprise cluster (P12d)

Default knowledge compose ([deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml)) runs a single **Community** `neo4j:5` node. For HA testing, use the Enterprise **3-primary** overlay.

## Compose

```bash
export NEO4J_ACCEPT_LICENSE_AGREEMENT=yes
docker compose \
  -f deploy/knowledge/compose.yml \
  -f deploy/knowledge/compose.neo4j-cluster.yml \
  up -d neo4j-core1 neo4j-core2 neo4j-core3
```

The overlay:

- Disables the single `neo4j` service (`profile: neo4j-single`, not started by default).
- Starts `neo4j-core{1,2,3}` on `neo4j:5-enterprise` with discovery v2 and separate data volumes.
- Sets `NEO4J_CLUSTER=1` and `NEO4J_URI=neo4j+routing://neo4j-core1:7687` on graph-bootstrap, ingest_worker, api, and mcp.

Optional host Bolt/Browser publish (dev): add `-f deploy/knowledge/compose.neo4j-publish.yml` and map ports on `neo4j-core1`.

## Application config

`knowledge/serve` reads cluster mode from the environment:

| Variable | Default (cluster) | Meaning |
|----------|-------------------|---------|
| `NEO4J_CLUSTER` | unset | `1` / `true` enables routing URI defaults |
| `NEO4J_URI` | `neo4j+routing://neo4j-core1:7687` when cluster | Bolt routing URI |
| `NEO4J_ROUTING_URI` | same as above | Used when `NEO4J_URI` is unset and cluster is on |

Community / local dev without `NEO4J_CLUSTER` keeps `neo4j://localhost:7687`.

## Smoke

Requires Enterprise license acceptance (no pull without it):

```bash
NEO4J_ACCEPT_LICENSE_AGREEMENT=yes ./scripts/test/smoke-neo4j-cluster.sh --up
NEO4J_ACCEPT_LICENSE_AGREEMENT=yes ./scripts/test/smoke-neo4j-cluster.sh
./scripts/test/smoke-neo4j-cluster.sh --down
```

Without `NEO4J_ACCEPT_LICENSE_AGREEMENT=yes`, the script exits 0 with `SKIP`.

## Related

- Graph read smoke (single node): `./scripts/test/smoke-graph-read.sh`
- [docs/architecture/threatintel-runtime.md](../architecture/threatintel-runtime.md) — default Neo4j service table
- P12 unified access ADR (when merged): `docs/architecture/platform-unified-access.md`
