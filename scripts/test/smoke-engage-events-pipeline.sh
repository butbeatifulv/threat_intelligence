#!/usr/bin/env bash
# Smoke: engage tool run -> engage.events -> bridge -> ingest.engage.* -> Neo4j (graph-ingest profile)
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(dirname "$0")/lib/smoke.sh"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

smoke_skip_no_docker

COMPOSE=(docker compose
  -f deploy/engage/compose.yml
  -f deploy/engage/compose.events.yml
)
PROFILES=(--profile graph-ingest)
ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:8890}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
WAIT_API_SEC="${SMOKE_EVENTS_API_WAIT_SEC:-120}"
WAIT_NEO4J_SEC="${SMOKE_EVENTS_NEO4J_WAIT_SEC:-90}"
POLL_INGEST_SEC="${SMOKE_EVENTS_INGEST_POLL_SEC:-60}"

log() { printf '[engage-events] %s\n' "$*"; }
fail() {
  log "FAIL: $*"
  "${COMPOSE[@]}" "${PROFILES[@]}" ps 2>/dev/null || true
  for svc in engage-api engage-events-worker ingest_worker neo4j nats; do
    log "--- logs ${svc} (tail 40) ---"
    "${COMPOSE[@]}" "${PROFILES[@]}" logs --tail=40 "${svc}" 2>/dev/null || true
  done
  exit 1
}

wait_neo4j() {
  local i=0
  while (( i < WAIT_NEO4J_SEC )); do
    if "${COMPOSE[@]}" "${PROFILES[@]}" exec -T neo4j \
      cypher-shell -u "${NEO4J_USER}" -p "${NEO4J_PASS}" "RETURN 1" >/dev/null 2>&1; then
      log "neo4j ready"
      return 0
    fi
    sleep 2
    i=$((i + 2))
  done
  fail "timeout waiting for neo4j (${WAIT_NEO4J_SEC}s)"
}

neo4j_tool_run_count() {
  "${COMPOSE[@]}" "${PROFILES[@]}" exec -T neo4j \
    cypher-shell -u "${NEO4J_USER}" -p "${NEO4J_PASS}" \
    "MATCH (r:EngageToolRun) RETURN count(r) AS c" 2>/dev/null | grep -Eo '[0-9]+' | tail -1 || echo 0
}

compose_down() {
  "${COMPOSE[@]}" "${PROFILES[@]}" down -v 2>/dev/null || true
}

log "starting events pipeline stack..."
"${COMPOSE[@]}" "${PROFILES[@]}" up -d --build \
  nats neo4j engage-api engage-events-worker ingest_worker \
  || fail "compose up failed (events + graph-ingest profile)"

trap compose_down EXIT

smoke_wait_http "${ENGAGE_URL}/health" "${WAIT_API_SEC}" "engage-api" 2 && log "engage-api ready" \
  || fail "timeout waiting for engage-api (${ENGAGE_URL}/health, ${WAIT_API_SEC}s)"
wait_neo4j

# Distroless API has no scanner binaries; runner emits audit even when binary is missing.
resp=$(curl -sS -w '\n%{http_code}' -X POST "${ENGAGE_URL}/api/tools/httpx_probe" \
  -H 'Content-Type: application/json' \
  -d '{"target":"https://example.com"}' 2>/dev/null || echo -e '\n000')
code=$(echo "${resp}" | tail -1)
body=$(echo "${resp}" | sed '$d')
if [[ "${code}" != "200" ]]; then
  fail "tool POST httpx_probe HTTP ${code}: ${body}"
fi
log "tool POST httpx_probe HTTP 200 (audit event expected even if success=false)"

deadline=$((SECONDS + POLL_INGEST_SEC))
count=0
while (( SECONDS < deadline )); do
  count=$(neo4j_tool_run_count)
  if [[ "${count}" =~ ^[0-9]+$ ]] && [[ "${count}" -ge 1 ]]; then
    log "EngageToolRun nodes in Neo4j: ${count}"
    echo "OK engage-events-pipeline smoke"
    exit 0
  fi
  sleep 2
done

fail "expected EngageToolRun count >= 1 in Neo4j after ${POLL_INGEST_SEC}s poll, got '${count}'"
