#!/usr/bin/env bash
# Smoke: CTF web workflow via engage-api.
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"

log() { printf '[ctf-web-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! smoke_wait_http "${ENGAGE_URL}/health" 5 "engage-api" 1 2>/dev/null; then
  log "SKIP: engage-api not reachable at ${ENGAGE_URL}"
  exit 0
fi

resp=$(curl -sf -X POST "${ENGAGE_URL}/api/ctf/create-challenge-workflow" \
  -H 'Content-Type: application/json' \
  -d '{"name":"smoke-web","category":"web","description":"sql injection test","target":"https://example.com"}')
echo "${resp}" | grep -q '"success":true' || fail "create-workflow failed: ${resp}"
log "OK create-challenge-workflow"

resp2=$(curl -sf -X POST "${ENGAGE_URL}/api/ctf/auto-solve-challenge" \
  -H 'Content-Type: application/json' \
  -d '{"name":"smoke-web","category":"web","target":"https://example.com","execute_tools":false}')
echo "${resp2}" | grep -q '"solve_result"' || fail "auto-solve failed: ${resp2}"
log "OK auto-solve-challenge (plan mode)"

log "ctf web smoke passed"
