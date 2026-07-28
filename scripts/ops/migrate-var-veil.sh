#!/usr/bin/env bash
# One-time migration: data/cache and data/neo4j_user_export → var/veil/
# Usage: ./scripts/ops/migrate-var-veil.sh [--dry-run]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
fi

log() { printf '[migrate-var-veil] %s\n' "$*"; }

mkdir -p "${SCRAPE_BLOB_DIR}" "${CRAWL_LEDGER_DATA_DIR}" "${PACK_RELEASES_DIR}"

migrate_dir() {
  local src="$1" dst="$2" label="$3"
  if [[ ! -d "${src}" ]]; then
    log "skip ${label}: ${src} not found"
    return 0
  fi
  local src_size dst_size
  src_size="$(du -sk "${src}" 2>/dev/null | awk '{print $1}')"
  log "${label}: rsync ${src} → ${dst} (~${src_size}K)"
  if [[ "$DRY_RUN" -eq 1 ]]; then
    return 0
  fi
  rsync -a "${src}/" "${dst}/"
  dst_size="$(du -sk "${dst}" 2>/dev/null | awk '{print $1}')"
  log "${label}: done (dest ~${dst_size}K)"
}

migrate_dir "${VEIL_ROOT}/data/cache" "${SCRAPE_BLOB_DIR}" "blobs"

migrate_graph_legacy() {
  local src="$1" label="$2"
  if [[ ! -d "${src}" ]]; then
    return 0
  fi
  if [[ -r "${src}" ]]; then
    migrate_dir "${src}" "${GRAPH_PACK_DIR}" "${label}"
    return 0
  fi
  log "warn: ${src} not readable (try: sudo rsync -a ${src}/ ${GRAPH_PACK_DIR}/)"
  if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    log "${label}: sudo rsync ${src} → ${GRAPH_PACK_DIR}"
    if [[ "$DRY_RUN" -eq 0 ]]; then
      sudo rsync -a "${src}/" "${GRAPH_PACK_DIR}/"
    fi
  fi
}

migrate_graph_legacy "${VEIL_ROOT}/data/neo4j_user_export" "graph"
migrate_graph_legacy "${VEIL_ROOT}/data/neo4j_export" "graph (neo4j_export)"

if [[ "$DRY_RUN" -eq 1 ]]; then
  log "dry-run: not clearing legacy data/"
  exit 0
fi

for legacy in data/cache data/neo4j_user_export data/neo4j_export; do
  if [[ -d "${VEIL_ROOT}/${legacy}" ]]; then
    find "${VEIL_ROOT}/${legacy}" -mindepth 1 -delete 2>/dev/null || rm -rf "${VEIL_ROOT:?}/${legacy:?}"/* 2>/dev/null || true
    touch "${VEIL_ROOT}/${legacy}/.gitkeep" 2>/dev/null || true
    log "cleared ${legacy}/ (kept .gitkeep if possible)"
  fi
done

log "migration complete — use var/veil/ for compose mounts"
