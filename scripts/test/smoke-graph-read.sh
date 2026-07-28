#!/usr/bin/env bash
# Graph read-path smoke: Neo4j + API (+ optional MCP HTTP). No scrape/NATS/ingest.
# Usage: ./scripts/test/smoke-graph-read.sh [--up] [--down]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

COMPOSE_FILES="${COMPOSE_FILES:--f deploy/knowledge/compose.yml -f deploy/knowledge/compose.graph-read.yml}"
API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
MCP_URL="${MCP_URL:-http://127.0.0.1:${MCP_HTTP_PORT:-8091}}"
WAIT_SEC="${SMOKE_GRAPH_WAIT_SEC:-300}"
WITH_MCP="${SMOKE_MCP:-1}"

compose_graph() {
  local -a prof=()
  [[ "${WITH_MCP}" == "1" ]] && prof=(--profile mcp)
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "${prof[@]}" "$@")
}

log() { printf '[graph-read-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if [[ "${1:-}" == "--down" ]]; then
  compose_graph down -v --remove-orphans 2>/dev/null || true
  exit 0
fi

if [[ "${1:-}" == "--up" ]]; then
  log "starting graph read stack (GRAPH_PACK_SKIP=1, no ingest)..."
  compose_graph down -v --remove-orphans 2>/dev/null || true
  compose_graph up -d --build neo4j graph-bootstrap api
  if [[ "${WITH_MCP}" == "1" ]]; then
    compose_graph up -d --build mcp
  fi
  shift
fi

wait_healthy() {
  local svc="$1"
  log "waiting for ${svc} healthy (max ${WAIT_SEC}s)..."
  local deadline=$((SECONDS + WAIT_SEC))
  while (( SECONDS < deadline )); do
    if compose_graph ps "$svc" 2>/dev/null | grep -q '(healthy)'; then
      log "${svc} healthy"
      return 0
    fi
    sleep 3
  done
  compose_graph ps "$svc" 2>/dev/null || true
  compose_graph logs --tail=40 "$svc" 2>/dev/null || true
  fail "${svc} not healthy in ${WAIT_SEC}s"
}

wait_healthy api
if [[ "${WITH_MCP}" == "1" ]]; then
  wait_healthy mcp
fi

curl_ok() {
  local url="$1"
  local expect="${2:-200}"
  local code
  code=$(curl -fsS -o /dev/null -w '%{http_code}' "$url" || echo "000")
  if [[ "$code" != "$expect" ]]; then
    fail "GET $url => $code (expected $expect)"
  fi
  log "OK $url ($code)"
}

curl_ok "${API_URL}/health"
curl_ok "${API_URL}/v1/categories"

if [[ "${WITH_MCP}" == "1" ]]; then
  curl_ok "${MCP_URL}/health"
  init_body='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}'
  init_resp=$(curl -fsS -X POST "${MCP_URL}/mcp" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "$init_body")
  echo "$init_resp" | grep -q '"protocolVersion"' || fail "MCP initialize missing protocolVersion"
  log "OK MCP initialize"

  list_body='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  list_resp=$(curl -fsS -X POST "${MCP_URL}/mcp" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "$list_body")
  echo "$list_resp" | grep -q 'ti_health' || fail "MCP tools/list missing ti_health"
  log "OK MCP tools/list"

  health_body='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"ti_health","arguments":{}}}'
  health_resp=$(curl -fsS -X POST "${MCP_URL}/mcp" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "$health_body")
  echo "$health_resp" | grep -q '"result"' || fail "MCP ti_health failed: $health_resp"
  log "OK MCP ti_health"
fi

log "graph read smoke passed"
