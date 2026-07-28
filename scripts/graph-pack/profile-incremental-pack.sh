#!/usr/bin/env bash
# Incremental graph pack: import BASE_GRAPH_PACK_VERSION into Neo4j, delta crawl using persistent ledger/blobs.
# Usage: ./scripts/graph-pack/profile-incremental-pack.sh [--no-down]
# Env: BASE_GRAPH_PACK_VERSION (default v0.4.1), place ZIP under var/veil/graph/releases/ for offline import.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

mkdir -p "${SCRAPE_BLOB_DIR}" "${CRAWL_LEDGER_DATA_DIR}" "${PACK_RELEASES_DIR}"

log() { printf '[graph-pack-incremental] %s\n' "$*"; }

if [[ "${1:-}" != "--no-down" ]]; then
  "${VEIL_ROOT}/scripts/ops/compose-down-ephemeral.sh"
fi

source_profile incremental-pack

base_zip="${PACK_RELEASES_DIR}/$(pack_zip_name "${BASE_GRAPH_PACK_VERSION}")"
if [[ -f "${base_zip}" ]]; then
  log "baseline pack on disk: ${base_zip}"
else
  log "baseline pack not local; graph-bootstrap will download veil-graph-${BASE_GRAPH_PACK_VERSION}"
fi

log "starting full stack (BASE=${BASE_GRAPH_PACK_VERSION}, sources=${SCRAPE_SOURCES}, FORCE_REFETCH=${SCRAPE_FORCE_REFETCH:-0})..."
exec "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
