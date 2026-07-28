## Veil graph pack ${GRAPH_PACK_VERSION}

Neo4j **5.x** export for [Veil](https://github.com/butbeautifulv/veil). Manifest schema: **`veil.graph-pack/1`**.

### Contents

- **ZIP:** `veil-graph-${GRAPH_PACK_VERSION}.zip`
- **Tag:** `veil-graph-${GRAPH_PACK_VERSION}`
- **Build profile:** ${BUILD_PROFILE:-fast-rich}
- **SHA256:** see `manifest.json` inside the archive

### Import

1. **Compose bootstrap** (default download): set `GRAPH_PACK_DEFAULT_URL` or use `graph-bootstrap` with `GRAPH_PACK_DEFAULT=1`.
2. **Local ZIP:** `GRAPH_PACK_FILE=/path/to/veil-graph-${GRAPH_PACK_VERSION}.zip` or bind-mount via [docker-compose.testpack.yml](../../docker-compose.testpack.yml).
3. **Script:** `./scripts/graph-pack/import.sh` with the ZIP path.

Env reference: [docs/architecture/threatintel-runtime.md](../threatintel-runtime.md#graph-bootstrap-usage-mode).

### Ingest changes since previous pack

${INGEST_CHANGELOG}

### Approximate graph size

${NODE_COUNTS:-Fill in after build (e.g. from Neo4j `MATCH (n) RETURN labels(n), count(*)`).}
