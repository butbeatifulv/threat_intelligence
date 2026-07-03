# Observability

Veil exposes complementary observability layers for the graph read plane and async workers:

| Layer | Tool | What it captures |
|-------|------|------------------|
| **HTTP RED metrics** | Prometheus | `veil_http_requests_total`, latency histograms on `veil-api` / `veil-mcp` |
| **Worker metrics** | Prometheus | NATS message rate/errors, Neo4j operation counters |
| **Traces** | Tempo (OpenTelemetry) | HTTP requests, worker jobs, NATS handoffs, Neo4j queries |
| **Structured logs** | Loki (Promtail) | JSON stdout with `service`, `trace_id`, `span_id` |

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `false` | Export traces to OTLP |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | Tempo OTLP gRPC endpoint |
| `OTEL_SERVICE_NAME` | per binary | Service name in Tempo |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Trace sampling ratio (1.0 = all spans offline) |
| `METRICS_LISTEN` | `:9090` | Worker metrics HTTP port |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `LOG_LEVEL` | `info` | slog level |

Helm values under `observability.*` in [deploy/helm/veil/values.yaml](../deploy/helm/veil/values.yaml).

## Local development

```bash
# Optional: run Tempo from cxado obs stack
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_SERVICE_NAME=veil-api

cd projects/veil
make test-knowledge   # after code changes
```

Compose profiles set `API_ENV=prod` / `MCP_ENV=prod` for JSON logs.

## k3s offline

```bash
VEIL_OFFLINE_TAG=offline-YYYYMMDD ./scripts/k8s/k3s-deploy-veil-offline.sh

# Optional workers overlay for pipeline E2E
VEIL_OFFLINE_TAG=offline-YYYYMMDD ./scripts/k8s/k3s-deploy-veil-offline.sh --with-workers-obs

./scripts/k8s/smoke-test-veil-obs.sh
```

Obs stack (Prometheus, Grafana, Tempo, Loki) lives in `cxado-obs` namespace — see [docs/deploy/k3s-offline-observability-optional.md](../../../docs/deploy/k3s-offline-observability-optional.md).

## Grafana

| Dashboard | UID | Content |
|-----------|-----|---------|
| Veil Graph | `veil-graph` | HTTP RED metrics |
| Veil Observability | `veil-observability` | Health, workers, logs, traces |

Explore shortcuts:

- Loki: `{namespace="veil", app="veil-api"} | json`
- Tempo: `{ resource.service.name = "veil-mcp" }`

## Cross-service traces (egregore → veil-mcp)

Egregore injects W3C `traceparent` via [`veil_mcp_client.py`](../../../egregore/cys_core/integrations/veil_mcp_client.py) (`inject_correlation_headers` + OTEL propagate).

Veil MCP HTTP uses `otelhttp` middleware — inbound calls continue the parent trace in Tempo.

**Verify:** run one egregore tool call to veil-mcp; Tempo service map should show `egregore-api` → `veil-mcp`.

## Code layout

| Package | Role |
|---------|------|
| [`pkg/observability/`](../pkg/observability/) | OTEL bootstrap, slog trace fields, worker metrics, NATS propagation |
| [`pkg/natsjet/`](../pkg/natsjet/) | W3C inject/extract on JetStream messages |
| [`knowledge/connector/neo4j/`](../knowledge/connector/neo4j/) | Neo4j query spans + metrics |

## Alerting

Prometheus rules: [deploy/observability/prometheus/rules/veil-alerts.yml](../../../deploy/observability/prometheus/rules/veil-alerts.yml)

Grafana unified alerting: [deploy/observability/grafana/provisioning/alerting/veil-alerts.yaml](../../../deploy/observability/grafana/provisioning/alerting/veil-alerts.yaml)
