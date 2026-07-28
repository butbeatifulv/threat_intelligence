#!/usr/bin/env bash
# Platform v3 P3: closed-loop pilot for target class "web host".
# Act (engage tool) -> learn (engage.events -> ingest.engage.*) -> remember (Neo4j) -> decide (veil-api + engage target-graph).
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(dirname "$0")/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

smoke_skip_no_docker

export COMPOSE_FILES="${VEIL_COMPOSE_FILES} -f deploy/engage/compose.yml -f deploy/engage/compose.veil-stack.yml"
export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
export SMOKE_VEIL_API_WAIT_SEC="${SMOKE_VEIL_API_WAIT_SEC:-300}"
export SMOKE_VEIL_ENGAGE_WAIT_SEC="${SMOKE_VEIL_ENGAGE_WAIT_SEC:-180}"
export SMOKE_VEIL_VEIL_API_WAIT_SEC="${SMOKE_VEIL_VEIL_API_WAIT_SEC:-180}"
export ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"
export API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
export SMOKE_ENGAGE_HOST="${SMOKE_ENGAGE_HOST:-example.com}"
export SMOKE_LOOP_POLL_SEC="${SMOKE_LOOP_POLL_SEC:-180}"

PROJECT="${COMPOSE_PROJECT_NAME:-platform-loop-$$}"
export COMPOSE_PROJECT_NAME="${PROJECT}"

log() { printf '[platform-closed-loop] %s\n' "$*"; }
fail() {
  log "FAIL: $*"
  compose ps 2>/dev/null || true
  for svc in engage-api api nats neo4j ingest_worker engage-events-worker; do
    log "--- logs ${svc} (tail 60) ---"
    compose logs --tail=60 "${svc}" 2>/dev/null || true
  done
  exit 1
}

cleanup() {
  compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

BUILD_FLAG=()
if [[ "${SMOKE_VEIL_STACK_BUILD:-1}" == "1" ]]; then
  BUILD_FLAG=(--build)
fi

log "pilot target class: web host (${SMOKE_ENGAGE_HOST})"
log "starting veil stack + engage overlay (project=${PROJECT})..."
compose up -d "${BUILD_FLAG[@]}" \
  nats neo4j ingest_worker api engage-api engage-events-worker

smoke_wait_http "${ENGAGE_URL}/health" "${SMOKE_VEIL_API_WAIT_SEC}" "engage-api" 2 \
  || fail "timeout engage-api (${ENGAGE_URL}/health, ${SMOKE_VEIL_API_WAIT_SEC}s)"
smoke_wait_http "${API_URL}/health" "${SMOKE_VEIL_VEIL_API_WAIT_SEC}" "veil-api" 2 \
  || fail "timeout veil-api (${API_URL}/health, ${SMOKE_VEIL_VEIL_API_WAIT_SEC}s)"
log "engage-api and veil-api healthy"

log "ACT: POST httpx_probe (audit -> engage.events -> ingest)"
code=$(curl -sS -o /tmp/platform-loop-tool.json -w '%{http_code}' -X POST "${ENGAGE_URL}/api/tools/httpx_probe" \
  -H 'Content-Type: application/json' \
  -d "{\"target\":\"https://${SMOKE_ENGAGE_HOST}\"}" 2>/dev/null || echo 000)
[[ "${code}" == "200" ]] || fail "tool POST HTTP ${code}"

log "REMEMBER: wait for engage nodes in veil-api search"
poll_deadline=$((SECONDS + SMOKE_LOOP_POLL_SEC))
found=0
engage_search_tmp="${TMPDIR:-/tmp}/platform-loop-engage-search-$$.json"
while (( SECONDS < poll_deadline )); do
  last_http=$(curl -sS -o "${engage_search_tmp}" -w '%{http_code}' \
    "${API_URL}/v1/categories/engage/search?q=${SMOKE_ENGAGE_HOST}&limit=10" 2>/dev/null || echo 000)
  if [[ "${last_http}" == "200" ]]; then
    if command -v jq >/dev/null 2>&1; then
      count=$(jq -r '((.nodes // []) | length)' "${engage_search_tmp}" 2>/dev/null || echo 0)
      if [[ "${count}" =~ ^[0-9]+$ ]] && [[ "${count}" -ge 1 ]]; then
        found=1
        log "veil-api engage search nodes=${count}"
        break
      fi
    elif grep -qi 'engagetoolrun\|engagefinding' "${engage_search_tmp}" 2>/dev/null; then
      found=1
      log "veil-api engage search returned engage nodes (jq not available)"
      break
    fi
  fi
  sleep 3
done
rm -f "${engage_search_tmp}" 2>/dev/null || true
[[ "${found}" -eq 1 ]] || fail "expected engage search hits for q=${SMOKE_ENGAGE_HOST} within ${SMOKE_LOOP_POLL_SEC}s"

ctx_http=$(curl -sS -o /tmp/platform-loop-veil-ctx.json -w '%{http_code}' \
  "${API_URL}/v1/categories/engage/context?q=${SMOKE_ENGAGE_HOST}" 2>/dev/null || echo 000)
[[ "${ctx_http}" == "200" ]] || fail "veil-api engage/context HTTP ${ctx_http}"
log "REMEMBER: veil-api engage/context HTTP 200"

log "DECIDE: engage target-graph read-back"
tg_http=$(curl -sS -o /tmp/platform-loop-target-graph.json -w '%{http_code}' \
  "${ENGAGE_URL}/api/intelligence/target-graph?target=https%3A%2F%2F${SMOKE_ENGAGE_HOST}" 2>/dev/null || echo 000)
[[ "${tg_http}" == "200" ]] || fail "target-graph HTTP ${tg_http}"
if command -v jq >/dev/null 2>&1; then
  enabled=$(jq -r '.graph_enabled // false' /tmp/platform-loop-target-graph.json 2>/dev/null || echo false)
  engage_nodes=$(jq -r '(.hits.engage.nodes // []) | length' /tmp/platform-loop-target-graph.json 2>/dev/null || echo 0)
  engage_found=$(jq -r '.engage_found // false' /tmp/platform-loop-target-graph.json 2>/dev/null || echo false)
  if [[ "${enabled}" != "true" ]] || { [[ "${engage_found}" != "true" ]] && [[ ! "${engage_nodes}" =~ ^[0-9]+$ || "${engage_nodes}" -lt 1 ]]; }; then
    cat /tmp/platform-loop-target-graph.json 2>/dev/null | head -c 2000 || true
    fail "target-graph missing engage memory (enabled=${enabled} nodes=${engage_nodes} found=${engage_found})"
  fi
  log "target-graph engage_nodes=${engage_nodes} engage_found=${engage_found}"
else
  grep -qi 'EngageToolRun\|engage_found\|"nodes"' /tmp/platform-loop-target-graph.json \
    || fail "target-graph body missing engage indicators"
fi

if command -v jq >/dev/null 2>&1; then
  tl_http=$(curl -sS -o /tmp/platform-loop-timeline.json -w '%{http_code}' \
    "${ENGAGE_URL}/api/intelligence/target-timeline?target=https%3A%2F%2F${SMOKE_ENGAGE_HOST}&limit=20" 2>/dev/null || echo 000)
  [[ "${tl_http}" == "200" ]] || fail "target-timeline HTTP ${tl_http}"
  audit_n=$(jq -r '(.audit_events // []) | length' /tmp/platform-loop-timeline.json 2>/dev/null || echo 0)
  log "LEARN: target-timeline audit_events=${audit_n}"
fi

log "OK platform closed-loop pilot (web host)"
exit 0
