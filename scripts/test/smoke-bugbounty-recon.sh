#!/usr/bin/env bash
# Smoke: bug bounty phased recon workflow via engage-api.
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"

log() { printf '[bb-recon-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! smoke_wait_http "${ENGAGE_URL}/health" 5 "engage-api" 1 2>/dev/null; then
  log "SKIP: engage-api not reachable at ${ENGAGE_URL}"
  exit 0
fi

resp=$(curl -sf -X POST "${ENGAGE_URL}/api/bugbounty/reconnaissance-workflow" \
  -H 'Content-Type: application/json' \
  -d '{"domain":"example.com"}')
echo "${resp}" | grep -q '"success":true' || fail "response: ${resp}"
echo "${resp}" | grep -q '"phases"' || fail "missing phases"
echo "${resp}" | grep -q 'subdomain_discovery' || fail "missing subdomain phase"
log "OK reconnaissance-workflow"

log "bug bounty recon smoke passed"
