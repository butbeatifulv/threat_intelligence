#!/usr/bin/env bash
# Platform v4 P4b: discover → enrich → remember → act → decide (minimal scrape + engage pilot).
# Heavy Docker smoke; default SKIP-friendly timeouts. Not for every PR.
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(dirname "$0")/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

smoke_skip_no_docker

export COMPOSE_FILES="${VEIL_COMPOSE_FILES} -f deploy/engage/compose.yml -f deploy/engage/compose.veil-stack.yml"
export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
export PIPELINE_WORKER_SCALE="${PIPELINE_WORKER_SCALE:-1}"
export INGEST_WORKER_SCALE="${INGEST_WORKER_SCALE:-1}"
export SCRAPE_WORKER_PARTITION="${SCRAPE_WORKER_PARTITION:-0}"
export SMOKE_SCRAPE_WAIT_SEC="${SMOKE_SCRAPE_WAIT_SEC:-900}"
export SMOKE_LOOP_POLL_SEC="${SMOKE_LOOP_POLL_SEC:-180}"
export ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"
export API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
export SMOKE_ENGAGE_HOST="${SMOKE_ENGAGE_HOST:-example.com}"

PROJECT="${COMPOSE_PROJECT_NAME:-platform-full-$$}"
export COMPOSE_PROJECT_NAME="${PROJECT}"

log() { printf '[platform-full-loop] %s\n' "$*"; }
fail() {
  log "FAIL: $*"
  compose ps 2>/dev/null || true
  exit 1
}

cleanup() {
  compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

source_profile smoke-minimal

BUILD_FLAG=()
if [[ "${SMOKE_VEIL_STACK_BUILD:-1}" == "1" ]]; then
  BUILD_FLAG=(--build)
fi

log "DISCOVER: full stack (profile=smoke-minimal) project=${PROJECT}"
PIPELINE_WORKER_SCALE="${PIPELINE_WORKER_SCALE}" \
  INGEST_WORKER_SCALE="${INGEST_WORKER_SCALE}" \
  SCRAPE_WORKER_PARTITION="${SCRAPE_WORKER_PARTITION}" \
  "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
compose up -d "${BUILD_FLAG[@]}" engage-api engage-events-worker

smoke_wait_http "${API_URL}/health" 300 "veil-api" 2 \
  || fail "timeout veil-api (${API_URL}/health, 300s)"
smoke_wait_http "${ENGAGE_URL}/health" 300 "engage-api" 2 \
  || fail "timeout engage-api (${ENGAGE_URL}/health, 300s)"
log "stack healthy"

log "waiting for scrape_worker exit (max ${SMOKE_SCRAPE_WAIT_SEC}s)..."
scrape_deadline=$((SECONDS + SMOKE_SCRAPE_WAIT_SEC))
scrape_ok=0
while (( SECONDS < scrape_deadline )); do
  if compose ps -a scrape_worker 2>/dev/null | grep -qE 'Exited \(0\)|exited \(0\)'; then
    scrape_ok=1
    break
  fi
  if compose ps -a scrape_worker 2>/dev/null | grep -qE 'Exited \([1-9]'; then
    fail "scrape_worker exited with error"
  fi
  sleep 10
done
[[ "${scrape_ok}" -eq 1 ]] || fail "scrape_worker did not exit 0 in time"

ioc_count=$(compose exec -T neo4j cypher-shell -u "${NEO4J_USER}" -p "${NEO4J_PASS}" \
  "MATCH (n:IOC) RETURN count(n) AS c" 2>/dev/null | grep -Eo '[0-9]+' | tail -1 || echo 0)
log "REMEMBER: Neo4j IOC count=${ioc_count}"
[[ "${ioc_count}" =~ ^[0-9]+$ ]] && [[ "${ioc_count}" -ge 1 ]] || fail "expected IOC nodes after minimal scrape"

log "ACT: engage httpx_probe"
code=$(curl -sS -o /tmp/platform-full-tool.json -w '%{http_code}' -X POST "${ENGAGE_URL}/api/tools/httpx_probe" \
  -H 'Content-Type: application/json' \
  -d "{\"target\":\"https://${SMOKE_ENGAGE_HOST}\"}" 2>/dev/null || echo 000)
[[ "${code}" == "200" ]] || fail "tool POST HTTP ${code}"

poll_deadline=$((SECONDS + SMOKE_LOOP_POLL_SEC))
found=0
while (( SECONDS < poll_deadline )); do
  http=$(curl -sS -o /tmp/platform-full-tg.json -w '%{http_code}' \
    "${ENGAGE_URL}/api/intelligence/target-graph?target=https%3A%2F%2F${SMOKE_ENGAGE_HOST}" 2>/dev/null || echo 000)
  if [[ "${http}" == "200" ]] && command -v jq >/dev/null 2>&1; then
    nodes=$(jq -r '(.hits.engage.nodes // []) | length' /tmp/platform-full-tg.json 2>/dev/null || echo 0)
    engage_found=$(jq -r '.engage_found // false' /tmp/platform-full-tg.json 2>/dev/null || echo false)
    if [[ "${engage_found}" == "true" ]] || [[ "${nodes}" =~ ^[0-9]+$ && "${nodes}" -ge 1 ]]; then
      found=1
      log "DECIDE: target-graph engage_nodes=${nodes} engage_found=${engage_found}"
      break
    fi
  fi
  sleep 3
done
[[ "${found}" -eq 1 ]] || fail "target-graph missing engage memory after act phase"

log "OK platform full loop (discover→enrich→remember→act→decide)"
exit 0
