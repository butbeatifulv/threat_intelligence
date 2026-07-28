#!/usr/bin/env bash
# Smoke: bug bounty recon with execute=true (phased tool runs).
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"
MAX_SEC="${BB_RECON_MAX_SEC:-900}"

log() { printf '[bb-recon-exec] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! smoke_wait_http "${ENGAGE_URL}/health" 5 "engage-api" 1 2>/dev/null; then
  log "SKIP: engage-api not reachable at ${ENGAGE_URL}"
  exit 0
fi

start=$(date +%s)
resp=$(curl -sf -X POST "${ENGAGE_URL}/api/bugbounty/reconnaissance-workflow" \
  -H 'Content-Type: application/json' \
  -d '{"domain":"example.com","execute":true}')
elapsed=$(( $(date +%s) - start ))

echo "${resp}" | grep -q '"success":true' || fail "response: ${resp}"
echo "${resp}" | grep -q 'phase_results' || fail "missing phase_results"
echo "${resp}" | grep -q 'tools_executed' || log "WARN: no tools_executed in phase_results (runner may lack binaries)"

if [[ "${elapsed}" -gt "${MAX_SEC}" ]]; then
  log "WARN: recon execute took ${elapsed}s (target < ${MAX_SEC}s)"
else
  log "timing: ${elapsed}s"
fi
log "OK reconnaissance-workflow execute"
