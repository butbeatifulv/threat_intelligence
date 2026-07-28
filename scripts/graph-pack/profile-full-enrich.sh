#!/usr/bin/env bash
# Full enrich crawl: incremental Neo4j seed + NVD_MAX_PAGES=10, ledger-aware.
# Usage: ./scripts/graph-pack/profile-full-enrich.sh [--skip-migrate] [--no-down]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

SKIP_MIGRATE=0
NO_DOWN=0
for arg in "$@"; do
  case "$arg" in
    --skip-migrate) SKIP_MIGRATE=1 ;;
    --no-down) NO_DOWN=1 ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

log() { printf '[graph-pack-full-enrich] %s\n' "$*"; }

if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
  "${VEIL_ROOT}/scripts/ops/migrate-var-veil.sh"
fi

mkdir -p "${SCRAPE_BLOB_DIR}" "${CRAWL_LEDGER_DATA_DIR}" "${PACK_RELEASES_DIR}"
chmod -R u+rwX "${VEIL_VAR_DIR}" 2>/dev/null || true

base_zip="${PACK_RELEASES_DIR}/$(pack_zip_name "${BASE_GRAPH_PACK_VERSION:-v0.4.1}")"
if [[ ! -f "${base_zip}" ]]; then
  log "downloading baseline pack $(pack_zip_name "${BASE_GRAPH_PACK_VERSION:-v0.4.1}")..."
  tmp_zip="$(mktemp)"
  curl -fsSL "$(pack_release_url "${BASE_GRAPH_PACK_VERSION:-v0.4.1}")" -o "${tmp_zip}"
  mv "${tmp_zip}" "${base_zip}"
fi

if [[ "$NO_DOWN" -eq 0 ]]; then
  "${VEIL_ROOT}/scripts/ops/compose-down-ephemeral.sh"
fi

source_profile full-enrich

log "baseline=${BASE_GRAPH_PACK_VERSION} NVD_MAX_PAGES=${NVD_MAX_PAGES} FORCE_REFETCH=${SCRAPE_FORCE_REFETCH:-0}"
exec "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
