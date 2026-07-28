#!/usr/bin/env bash
# Smoke: veil-api engage category read (search, context, optional GetNode by hostname).
# Prerequisite: veil-api up with Neo4j graph (ingest optional for non-empty hits).
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
HOST="${SMOKE_ENGAGE_HOST:-example.com}"

log() { printf '[graph-engage-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! smoke_wait_http "${API_URL}/health" 5 "veil-api" 1 2>/dev/null; then
  log "SKIP: veil-api not reachable at ${API_URL}"
  exit 0
fi

cats=$(curl -sf "${API_URL}/v1/categories" 2>/dev/null || echo '[]')
if ! echo "${cats}" | grep -q '"engage"'; then
  fail "GET /v1/categories missing engage category"
fi
log "OK categories includes engage"

code=$(curl -fsS -o /tmp/veil-engage-search.json -w '%{http_code}' \
  "${API_URL}/v1/categories/engage/search?q=${HOST}&limit=10" 2>/dev/null || echo "000")
if [[ "${code}" != "200" ]]; then
  fail "engage search HTTP ${code}"
fi
log "OK GET /v1/categories/engage/search?q=${HOST}"

ctx_code=$(curl -fsS -o /tmp/veil-engage-context.json -w '%{http_code}' \
  "${API_URL}/v1/categories/engage/context?q=${HOST}" 2>/dev/null || echo "000")
if [[ "${ctx_code}" != "200" ]]; then
  fail "engage context HTTP ${ctx_code}"
fi
log "OK GET /v1/categories/engage/context?q=${HOST}"

node_code=$(curl -fsS -o /tmp/veil-engage-node.json -w '%{http_code}' \
  "${API_URL}/v1/nodes/${HOST}" 2>/dev/null || echo "000")
if [[ "${node_code}" == "200" ]]; then
  log "OK GET /v1/nodes/${HOST} (EngageTarget by name)"
elif [[ "${node_code}" == "404" ]]; then
  log "GET /v1/nodes/${HOST} => 404 (no ingested target yet; lookup path OK)"
else
  fail "GET /v1/nodes/${HOST} => ${node_code}"
fi

log "graph engage category smoke passed"
