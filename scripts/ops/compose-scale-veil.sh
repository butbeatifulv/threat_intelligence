#!/usr/bin/env bash
# Scale stateless Veil graph API/MCP, Engage HTTP/MCP/worker, and NATS workers.
# Requires infra (crawl-db, nats, neo4j, graph-bootstrap) — run compose-up-full.sh first
# or set COMPOSE_SCALE_BRING_INFRA=1.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

VEIL_API_SCALE="${VEIL_API_SCALE:-1}"
VEIL_MCP_SCALE="${VEIL_MCP_SCALE:-1}"
ENGAGE_API_SCALE="${ENGAGE_API_SCALE:-1}"
ENGAGE_MCP_SCALE="${ENGAGE_MCP_SCALE:-1}"
ENGAGE_WORKER_SCALE="${ENGAGE_WORKER_SCALE:-1}"
PIPELINE_SCALE="${PIPELINE_WORKER_SCALE:-1}"
INGEST_SCALE="${INGEST_WORKER_SCALE:-1}"
BRING_INFRA="${COMPOSE_SCALE_BRING_INFRA:-0}"
WITH_ENGAGE="${COMPOSE_SCALE_ENGAGE:-1}"

log() { printf '[compose-scale-veil] %s\n' "$*"; }

if [[ "$BRING_INFRA" == "1" ]]; then
  log "bringing up infrastructure (crawl-db, nats, neo4j, graph-bootstrap)..."
  compose_with_extras up -d --build crawl-db nats neo4j graph-bootstrap
fi

graph_scale=(
  --scale "api=${VEIL_API_SCALE}"
  --scale "pipeline_worker=${PIPELINE_SCALE}"
  --scale "ingest_worker=${INGEST_SCALE}"
)
graph_services=(api pipeline_worker ingest_worker)

if [[ "${VEIL_MCP_SCALE}" != "0" ]]; then
  export COMPOSE_PROFILES="${COMPOSE_PROFILES:-mcp}"
  graph_scale+=(--scale "mcp=${VEIL_MCP_SCALE}")
  graph_services+=(mcp)
fi

log "graph tier (api=${VEIL_API_SCALE}, mcp=${VEIL_MCP_SCALE}, pipeline=${PIPELINE_SCALE}, ingest=${INGEST_SCALE})..."
compose_with_extras up -d --build "${graph_scale[@]}" "${graph_services[@]}"

if [[ "$WITH_ENGAGE" == "1" ]]; then
  export COMPOSE_FILES="${VEIL_COMPOSE_FILES} -f deploy/engage/compose.yml"
  engage_scale=(
    --scale "engage-api=${ENGAGE_API_SCALE}"
    --scale "engage-mcp=${ENGAGE_MCP_SCALE}"
    --scale "engage-worker=${ENGAGE_WORKER_SCALE}"
  )
  log "engage tier (api=${ENGAGE_API_SCALE}, mcp=${ENGAGE_MCP_SCALE}, worker=${ENGAGE_WORKER_SCALE})..."
  compose up -d --build "${engage_scale[@]}" engage-api engage-mcp engage-worker
fi

log "done — stateless replicas applied (host :8090/:8890 conflict when scale>1 without edge nginx)"
