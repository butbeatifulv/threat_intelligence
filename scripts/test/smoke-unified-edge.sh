#!/usr/bin/env bash
# P12i: unified platform nginx edge — graph /v1, engage /api, MCP /mcp/graph + /mcp/engage.
# Usage: ./scripts/test/smoke-unified-edge.sh [--up] [--down]
# Env: SMOKE_SKIP_UNIFIED_EDGE=1, SMOKE_UNIFIED_EDGE_BUILD=1, VEIL_EDGE_HTTPS_PORT, GRAPH_PACK_SKIP=1
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(dirname "$0")/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

if [[ "${SMOKE_SKIP_UNIFIED_EDGE:-0}" == "1" ]]; then
  echo "[unified-edge-smoke] SKIP: SMOKE_SKIP_UNIFIED_EDGE=1" >&2
  exit 0
fi

smoke_skip_no_docker

export COMPOSE_FILES="-f deploy/knowledge/compose.yml -f deploy/knowledge/compose.graph-read.yml -f deploy/engage/compose.yml -f deploy/platform/compose.edge.yml -f deploy/platform/compose.edge-hostless.yml"
export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
export VEIL_EDGE_HTTPS_PORT="${VEIL_EDGE_HTTPS_PORT:-8443}"
export EDGE_URL="https://127.0.0.1:${VEIL_EDGE_HTTPS_PORT}"
export WAIT_SEC="${SMOKE_UNIFIED_EDGE_WAIT_SEC:-420}"
export CURL=(curl -skS --connect-timeout 5 --max-time 30)

CERT_DIR="${VEIL_ROOT}/deploy/platform/nginx/certs"
CERT_CRT="${VEIL_EDGE_TLS_CERT:-${CERT_DIR}/tls.crt}"
CERT_KEY="${VEIL_EDGE_TLS_KEY:-${CERT_DIR}/tls.key}"

PROJECT="${COMPOSE_PROJECT_NAME:-unified-edge-smoke-$$}"
export COMPOSE_PROJECT_NAME="${PROJECT}"

log() { printf '[unified-edge-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; compose ps 2>/dev/null || true; exit 1; }

ensure_tls() {
  if [[ -f "${CERT_CRT}" && -f "${CERT_KEY}" ]]; then
    return 0
  fi
  log "generating dev TLS in ${CERT_DIR}..."
  mkdir -p "${CERT_DIR}"
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "${CERT_KEY}" \
    -out "${CERT_CRT}" \
    -subj '/CN=localhost'
}

compose_edge() {
  local -a prof=(--profile mcp)
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "${prof[@]}" "$@")
}

if [[ "${1:-}" == "--down" ]]; then
  compose_edge down -v --remove-orphans 2>/dev/null || true
  exit 0
fi

cleanup() {
  if [[ "${SMOKE_UNIFIED_EDGE_NO_CLEANUP:-0}" != "1" ]]; then
    compose_edge down -v --remove-orphans 2>/dev/null || true
  fi
}

if [[ "${1:-}" != "--up" ]]; then
  trap cleanup EXIT
fi

wait_healthy() {
  local svc="$1"
  log "waiting for ${svc} healthy (max ${WAIT_SEC}s)..."
  local deadline=$((SECONDS + WAIT_SEC))
  while (( SECONDS < deadline )); do
    if compose_edge ps "$svc" 2>/dev/null | grep -q '(healthy)'; then
      log "${svc} healthy"
      return 0
    fi
    sleep 3
  done
  compose_edge logs --tail=60 "$svc" 2>/dev/null || true
  fail "${svc} not healthy in ${WAIT_SEC}s"
}

curl_code() {
  local url="$1"
  "${CURL[@]}" -o /dev/null -w '%{http_code}' "$url" 2>/dev/null || echo "000"
}

mcp_post() {
  local url="$1"
  local body="$2"
  "${CURL[@]}" -X POST "$url" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "$body"
}

if [[ "${1:-}" == "--up" ]]; then
  ensure_tls
  export VEIL_EDGE_TLS_CERT="${CERT_CRT}"
  export VEIL_EDGE_TLS_KEY="${CERT_KEY}"
  log "starting unified-edge stack (project=${PROJECT})..."
  compose_edge down -v --remove-orphans 2>/dev/null || true
  BUILD_FLAG=()
  [[ "${SMOKE_UNIFIED_EDGE_BUILD:-1}" == "1" ]] && BUILD_FLAG=(--build)
  compose_edge up -d "${BUILD_FLAG[@]}" \
    neo4j graph-bootstrap api mcp engage-api engage-mcp veil-edge
  shift
  log "stack up; run without --up to execute curl checks"
  exit 0
fi

ensure_tls
export VEIL_EDGE_TLS_CERT="${CERT_CRT}"
export VEIL_EDGE_TLS_KEY="${CERT_KEY}"

log "starting unified-edge stack (project=${PROJECT}, GRAPH_PACK_SKIP=${GRAPH_PACK_SKIP})..."
compose_edge down -v --remove-orphans 2>/dev/null || true
BUILD_FLAG=()
[[ "${SMOKE_UNIFIED_EDGE_BUILD:-1}" == "1" ]] && BUILD_FLAG=(--build)
compose_edge up -d "${BUILD_FLAG[@]}" \
  neo4j graph-bootstrap api mcp engage-api engage-mcp veil-edge

wait_healthy api
wait_healthy mcp
wait_healthy engage-api
wait_healthy engage-mcp

wait_edge() {
  log "waiting for veil-edge at ${EDGE_URL}/health (max 120s)..."
  local deadline=$((SECONDS + 120))
  while (( SECONDS < deadline )); do
    if [[ "$(curl_code "${EDGE_URL}/health")" == "200" ]]; then
      log "veil-edge reachable"
      return 0
    fi
    sleep 2
  done
  compose_edge logs --tail=40 veil-edge 2>/dev/null || true
  fail "veil-edge not reachable at ${EDGE_URL}/health"
}
wait_edge

code=$(curl_code "${EDGE_URL}/health")
[[ "$code" == "200" ]] || fail "GET ${EDGE_URL}/health => ${code}"
log "OK ${EDGE_URL}/health"

code=$(curl_code "${EDGE_URL}/v1/categories")
[[ "$code" == "200" ]] || fail "GET ${EDGE_URL}/v1/categories => ${code}"
log "OK ${EDGE_URL}/v1/categories"

code=$(curl_code "${EDGE_URL}/api/tools")
[[ "$code" == "200" ]] || fail "GET ${EDGE_URL}/api/tools => ${code}"
log "OK ${EDGE_URL}/api/tools"

init_body='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}'
graph_init=$(mcp_post "${EDGE_URL}/mcp/graph/" "$init_body")
echo "$graph_init" | grep -q '"protocolVersion"' || fail "graph MCP initialize: ${graph_init}"
log "OK ${EDGE_URL}/mcp/graph/ initialize"

list_body='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
graph_list=$(mcp_post "${EDGE_URL}/mcp/graph/" "$list_body")
echo "$graph_list" | grep -q 'ti_health' || fail "graph MCP tools/list missing ti_health"
log "OK ${EDGE_URL}/mcp/graph/ tools/list"

engage_init=$(mcp_post "${EDGE_URL}/mcp/engage/" "$init_body")
echo "$engage_init" | grep -q '"protocolVersion"' || fail "engage MCP initialize: ${engage_init}"
log "OK ${EDGE_URL}/mcp/engage/ initialize"

engage_list=$(mcp_post "${EDGE_URL}/mcp/engage/" "$list_body")
echo "$engage_list" | grep -q '"tools"' || fail "engage MCP tools/list missing tools"
log "OK ${EDGE_URL}/mcp/engage/ tools/list"

log "unified-edge smoke passed"
