#!/usr/bin/env bash
# Crawl ledger summary + blob store size. Requires crawl-db (or CRAWL_MYSQL DSN).
# Usage: ./scripts/crawl/status.sh
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[crawl-status] %s\n' "$*"; }

run_sql() {
  local q="$1"
  if compose ps crawl-db 2>/dev/null | grep -qE 'Up|running'; then
    compose exec -T crawl-db mysql -uveil -pveilpass veil_ledger -N -e "$q" 2>/dev/null
    return 0
  fi
  local dsn="${CRAWL_MYSQL:-veil:veilpass@tcp(127.0.0.1:3306)/veil_ledger}"
  if command -v mysql >/dev/null 2>&1; then
    mysql --protocol=TCP -h 127.0.0.1 -P 3306 -uveil -pveilpass veil_ledger -N -e "$q" 2>/dev/null
    return 0
  fi
  return 1
}

log "var/veil paths"
printf '  blobs:  %s\n' "${SCRAPE_BLOB_DIR}"
printf '  ledger: %s\n' "${CRAWL_LEDGER_DATA_DIR}"
printf '  graph:  %s\n' "${GRAPH_PACK_DIR}"

if [[ -d "${SCRAPE_BLOB_DIR}" ]]; then
  blob_size="$(du -sh "${SCRAPE_BLOB_DIR}" 2>/dev/null | awk '{print $1}' || echo '?')"
  blob_files="$(find "${SCRAPE_BLOB_DIR}" -type f 2>/dev/null | wc -l | tr -d ' ')"
  printf '  blob size: %s (%s files)\n' "${blob_size}" "${blob_files}"
else
  log "blob dir missing (will be created on first crawl)"
fi

if ! run_sql 'SELECT 1' >/dev/null 2>&1; then
  log "crawl-db not reachable — start stack or set CRAWL_MYSQL"
  exit 0
fi

log "crawl_resource by source"
run_sql "SELECT source, COUNT(*) AS n FROM crawl_resource GROUP BY source ORDER BY n DESC;"

log "recent fetches (last 10)"
run_sql "SELECT resource_key, source, last_fetched_at, LEFT(content_sha256,12) FROM crawl_resource ORDER BY last_fetched_at DESC LIMIT 10;"

log "static resources (PolicyStatic, fetched once)"
run_sql "SELECT COUNT(*) FROM crawl_resource WHERE fetch_policy='static';"

log "hint: incremental pack build → ./scripts/graph-pack/profile-incremental-pack.sh"
