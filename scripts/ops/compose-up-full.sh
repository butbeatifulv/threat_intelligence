#!/usr/bin/env bash
# Start full Veil stack (scrape + pipeline + graph) with optional worker scaling.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

PIPELINE_SCALE="${PIPELINE_WORKER_SCALE:-1}"
INGEST_SCALE="${INGEST_WORKER_SCALE:-1}"
SCRAPE_PARTITION="${SCRAPE_WORKER_PARTITION:-0}"

scale_args=(--scale "pipeline_worker=${PIPELINE_SCALE}" --scale "ingest_worker=${INGEST_SCALE}")

log() { printf '[compose-up-full] %s\n' "$*"; }

log "bringing up infrastructure (crawl-db, nats, neo4j, graph-bootstrap)..."
compose_with_extras up -d --build crawl-db nats neo4j graph-bootstrap

log "starting workers (pipeline_scale=${PIPELINE_SCALE}, ingest_scale=${INGEST_SCALE}, scrape_partition=${SCRAPE_PARTITION})..."
if [[ "$SCRAPE_PARTITION" == "1" ]]; then
  compose_with_extras up -d --build "${scale_args[@]}" \
    scrape_worker_fast scrape_worker_slow pipeline_worker ingest_worker
else
  compose_with_extras up -d --build "${scale_args[@]}" \
    scrape_worker pipeline_worker ingest_worker
fi

log "optional: api"
compose_with_extras up -d api 2>/dev/null || true

log "done — check: docker compose ps"
