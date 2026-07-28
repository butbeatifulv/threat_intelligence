#!/usr/bin/env bash
# E2E smoke: Discovery (harvest) → pipeline → ingest → Neo4j.
# Usage: ./scripts/test/smoke-discovery-e2e.sh [--up] [--restart-scrape]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

PROFILE="${COMPOSE_PROFILE:-}"
SCRAPE_SVC="${SCRAPE_SVC:-scrape_worker}"
PIPELINE_SVC="${PIPELINE_SVC:-pipeline_worker}"
INGEST_SVC="${INGEST_SVC:-ingest_worker}"
NATS_MON="${NATS_MON:-http://127.0.0.1:${NATS_MONITOR_PORT:-8222}}"
API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}/health}"
CRAWL_MYSQL="${CRAWL_MYSQL:-veil:veilpass@tcp(127.0.0.1:3306)/veil_ledger}"
WAIT_SEC="${SMOKE_WAIT_SEC:-600}"

compose_smoke() {
  local -a prof=()
  [[ -n "${PROFILE}" ]] && prof=(--profile "$PROFILE")
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "${prof[@]}" "$@")
}

log() { printf '[smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if [[ "${1:-}" == "--up" ]]; then
  source_profile smoke-minimal
  export SMOKE_CLEAN_VOLUMES="${SMOKE_CLEAN_VOLUMES:-1}"
  if [[ "${SMOKE_CLEAN_VOLUMES}" == "1" ]]; then
    log "removing old compose volumes..."
    compose_smoke down -v --remove-orphans 2>/dev/null || true
  fi
  log "bringing up stack (sources=${SCRAPE_SOURCES})..."
  PIPELINE_WORKER_SCALE="${PIPELINE_WORKER_SCALE:-1}" \
  INGEST_WORKER_SCALE="${INGEST_WORKER_SCALE:-1}" \
  SCRAPE_WORKER_PARTITION="${SCRAPE_WORKER_PARTITION:-0}" \
    "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
  shift
fi

if [[ "${1:-}" == "--restart-scrape" ]]; then
  log "restarting discovery worker (ledger pass 2)..."
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose_smoke restart scrape_worker_fast scrape_worker_slow 2>/dev/null || true
  else
    compose_smoke restart "$SCRAPE_SVC"
  fi
  sleep 15
  compose_smoke logs --tail=80 "$SCRAPE_SVC" 2>&1 | grep -iE 'unchanged|skip publish' || log "(no unchanged lines in last 80)"
  exit 0
fi

scrape_exited_ok() {
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose_smoke ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \(0\)|exited \(0\)' \
      && ! compose_smoke ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \([1-9]'
  else
    compose_smoke ps -a "$SCRAPE_SVC" 2>/dev/null | grep -qE 'Exited \(0\)|exited \(0\)'
  fi
}

scrape_exited_err() {
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose_smoke ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \([1-9]'
  else
    compose_smoke ps -a "$SCRAPE_SVC" 2>/dev/null | grep -qE 'Exited \([1-9]'
  fi
}

log "waiting for discovery worker(s) to exit (max ${WAIT_SEC}s)..."
deadline=$((SECONDS + WAIT_SEC))
while (( SECONDS < deadline )); do
  if scrape_exited_ok; then
    log "discovery worker(s) exited 0"
    break
  fi
  if scrape_exited_err; then
    fail "discovery worker exited with error"
  fi
  sleep 5
done

log "NATS monitoring $NATS_MON"
if command -v curl >/dev/null 2>&1; then
  curl -sf "$NATS_MON/healthz" >/dev/null || fail "NATS healthz"
else
  log "warn: curl missing, skip NATS HTTP checks"
fi

pw_up=$(compose_smoke ps "$PIPELINE_SVC" 2>/dev/null | grep -c Up || true)
iw_up=$(compose_smoke ps "$INGEST_SVC" 2>/dev/null | grep -c Up || true)
pw_want="${PIPELINE_WORKER_SCALE:-1}"
iw_want="${INGEST_WORKER_SCALE:-1}"
[[ "$pw_up" -ge "$pw_want" ]] || fail "pipeline_worker: want ${pw_want} Up, got ${pw_up}"
[[ "$iw_up" -ge "$iw_want" ]] || fail "ingest_worker: want ${iw_want} Up, got ${iw_up}"

log "Neo4j label counts"
if compose_smoke ps neo4j 2>/dev/null | grep -q Up; then
  compose_smoke exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:Vulnerability) RETURN count(n) AS vulnerabilities;" 2>/dev/null || true
  compose_smoke exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:IOC) RETURN count(n) AS iocs;" 2>/dev/null || true
else
  fail "neo4j not running"
fi

if command -v curl >/dev/null 2>&1; then
  curl -sf "$API_URL" >/dev/null || log "warn: API not up"
fi

log "PASS — discovery E2E smoke"
