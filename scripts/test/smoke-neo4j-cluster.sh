#!/usr/bin/env bash
# Neo4j Enterprise 3-primary cluster smoke (optional; requires license acceptance).
# Usage: NEO4J_ACCEPT_LICENSE_AGREEMENT=yes ./scripts/test/smoke-neo4j-cluster.sh [--up] [--down]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[neo4j-cluster-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if [[ "${NEO4J_ACCEPT_LICENSE_AGREEMENT:-}" != "yes" ]]; then
  log "SKIP: Neo4j Enterprise cluster smoke (set NEO4J_ACCEPT_LICENSE_AGREEMENT=yes)"
  exit 0
fi

COMPOSE_FILES="${COMPOSE_FILES:--f deploy/knowledge/compose.yml -f deploy/knowledge/compose.graph-read.yml -f deploy/knowledge/compose.neo4j-cluster.yml}"
WAIT_SEC="${SMOKE_NEO4J_CLUSTER_WAIT_SEC:-600}"
API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"

compose_cluster() {
  export NEO4J_ACCEPT_LICENSE_AGREEMENT=yes
  export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "$@")
}

if [[ "${1:-}" == "--down" ]]; then
  compose_cluster down -v --remove-orphans 2>/dev/null || true
  exit 0
fi

if [[ "${1:-}" == "--up" ]]; then
  log "starting Neo4j Enterprise cluster + graph read stack..."
  compose_cluster down -v --remove-orphans 2>/dev/null || true
  compose_cluster up -d --build neo4j-core1 neo4j-core2 neo4j-core3 graph-bootstrap api
  shift
fi

wait_healthy() {
  local svc="$1"
  log "waiting for ${svc} healthy (max ${WAIT_SEC}s)..."
  local deadline=$((SECONDS + WAIT_SEC))
  while (( SECONDS < deadline )); do
    if compose_cluster ps "$svc" 2>/dev/null | grep -q '(healthy)'; then
      log "${svc} healthy"
      return 0
    fi
    sleep 5
  done
  compose_cluster ps "$svc" 2>/dev/null || true
  compose_cluster logs --tail=60 "$svc" 2>/dev/null || true
  fail "${svc} not healthy in ${WAIT_SEC}s"
}

for core in neo4j-core1 neo4j-core2 neo4j-core3; do
  wait_healthy "$core"
done

wait_healthy api

show_servers=$(compose_cluster exec -T neo4j-core1 cypher-shell -u neo4j -p neo4jpassword \
  "SHOW SERVERS YIELD name, state, health" 2>/dev/null || true)
if [[ -z "$show_servers" ]]; then
  fail "SHOW SERVERS returned no output"
fi
echo "$show_servers" | grep -qi 'online\|enabled\|available' || fail "cluster members not online: $show_servers"
log "OK cluster members:\n$show_servers"

code=$(curl -fsS -o /dev/null -w '%{http_code}' "${API_URL}/health" || echo "000")
if [[ "$code" != "200" ]]; then
  fail "GET ${API_URL}/health => $code"
fi
log "OK ${API_URL}/health ($code)"

log "neo4j enterprise cluster smoke passed"
