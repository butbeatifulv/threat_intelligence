#!/usr/bin/env bash
# Smoke: CTF pwn/binary workflow via engage-api.
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"

log() { printf '[ctf-pwn-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! smoke_wait_http "${ENGAGE_URL}/health" 5 "engage-api" 1 2>/dev/null; then
  log "SKIP: engage-api not reachable at ${ENGAGE_URL}"
  exit 0
fi

resp=$(curl -sf -X POST "${ENGAGE_URL}/api/ctf/create-challenge-workflow" \
  -H 'Content-Type: application/json' \
  -d '{"name":"smoke-pwn","category":"pwn","description":"buffer overflow binary","difficulty":"medium"}')
echo "${resp}" | grep -q '"success":true' || fail "create-workflow: ${resp}"
log "OK create-challenge-workflow pwn"

resp2=$(curl -sf -X POST "${ENGAGE_URL}/api/ctf/suggest-tools" \
  -H 'Content-Type: application/json' \
  -d '{"description":"buffer overflow pwn","category":"pwn"}')
echo "${resp2}" | grep -q 'checksec\|suggested_tools' || fail "suggest-tools: ${resp2}"
log "OK suggest-tools pwn"

# binary-analyzer requires file_path; skip if no sample
if [[ -n "${CTF_SMOKE_BINARY:-}" ]] && [[ -f "${CTF_SMOKE_BINARY}" ]]; then
  curl -sf -X POST "${ENGAGE_URL}/api/ctf/binary-analyzer" \
    -H 'Content-Type: application/json' \
    -d "{\"binary_path\":\"${CTF_SMOKE_BINARY}\"}" | grep -q '"success":true' || fail "binary-analyzer"
  log "OK binary-analyzer"
else
  log "SKIP binary-analyzer (set CTF_SMOKE_BINARY to a file path)"
fi

log "ctf pwn smoke passed"
