# Veil Helm chart (control plane)

Kubernetes workloads: graph `api` / `mcp`, `engage-api` / `engage-mcp`, autoscaled `pipeline_worker` / `ingest_worker`, `scrape` CronJob.

```bash
helm template veil . -f values.yaml -f values-stage.yaml \
  --set global.imageTag=v0.4.5
```

Stateful services (Neo4j, NATS) default to Ansible+Compose data nodes — set `global.natsUrl` / `global.neo4jUri` accordingly. Prod Neo4j: Enterprise **3-core** cluster URI (`neo4j+routing://…`) per [docs/architecture/platform-unified-access.md](../../../docs/architecture/platform-unified-access.md).

## Replicas and HPA

| Workload | Values key | Scale notes |
|----------|------------|-------------|
| Graph API | `graph.api.replicas` | Stateless; raise for 4/8/16 HTTP capacity. Optional `graph.api.hpa` (CPU; disabled by default). |
| Graph MCP | `graph.mcp.replicas` | Stateless Streamable HTTP on `:8091`; same HPA pattern as `graph.api`. |
| Engage API | `engage.api.replicas` | Stateless; `engage.api.hpa` optional. |
| Engage MCP | `engage.mcp.replicas` | Stateless; `engage.mcp.hpa` optional. |
| Pipeline worker | `pipeline.worker.replicas` + `pipeline.worker.hpa` | JetStream **shared durable** — safe to scale out (competing consumers). |
| Ingest worker | `ingest.worker.replicas` + `ingest.worker.hpa` | Shared durable ingest consumer — safe to scale out. |
| Scrape | `scrape.cronJob` | **Do not** run duplicate CronJobs with the same `SCRAPE_SOURCES` profile; one scheduled job per source partition. |

Worker HPAs use CPU today; wire NATS consumer lag metrics in a later phase. Stateless API/MCP HPAs are off by default — prefer explicit `replicas` (e.g. `4`, `8`, `16`) until unified edge ingress and metrics are in place.

Compose equivalent: `scripts/ops/compose-scale-veil.sh` (`VEIL_API_SCALE`, `VEIL_MCP_GRAPH_SCALE`, `VEIL_ENGAGE_API_SCALE`, `VEIL_MCP_ENGAGE_SCALE`).

## Stateless scale presets (4 / 8 / 16)

**Stage (minimal):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-stage.yaml \
  --set graph.api.replicas=1 \
  --set engage.api.replicas=1 \
  --set pipeline.worker.replicas=1 \
  --set ingest.worker.replicas=1
```

**Medium (4 replicas):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=4 \
  --set engage.api.replicas=4 \
  --set pipeline.worker.replicas=4 \
  --set ingest.worker.replicas=4
```

**High (8 replicas):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=8 \
  --set engage.api.replicas=8 \
  --set pipeline.worker.replicas=8 \
  --set ingest.worker.replicas=8
```

**Peak (16 replicas):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=16 \
  --set engage.api.replicas=16 \
  --set pipeline.worker.replicas=16 \
  --set pipeline.worker.hpa.maxReplicas=16 \
  --set ingest.worker.replicas=16 \
  --set ingest.worker.hpa.maxReplicas=16
```

MCP HTTP deployments and unified Ingress path routing: [docs/architecture/platform-unified-access.md](../../../docs/architecture/platform-unified-access.md).
