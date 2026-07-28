#!/usr/bin/env bash
# Fast-rich graph pack profile (~25 min): all 7 sources, minimal NVD, no Atomic/MSF bulk.
# Usage: ./scripts/graph-pack/profile-fast-rich.sh [--no-down] [--full]
#   --full  wipe var/veil ledger+blobs and force refetch (legacy cold rebuild)
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[graph-pack-fast-rich] %s\n' "$*"; }

FULL=0
NO_DOWN=0
for arg in "$@"; do
  case "$arg" in
    --full) FULL=1 ;;
    --no-down) NO_DOWN=1 ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

mkdir -p "${SCRAPE_BLOB_DIR}" "${CRAWL_LEDGER_DATA_DIR}" "${PACK_RELEASES_DIR}"

if [[ "$NO_DOWN" -eq 0 ]]; then
  if [[ "$FULL" -eq 1 ]]; then
    log "full rebuild: stopping stack, wiping ephemeral volumes + var/veil ledger/blobs..."
    compose down -v --remove-orphans 2>/dev/null || true
    rm -rf "${CRAWL_LEDGER_DATA_DIR:?}"/* "${SCRAPE_BLOB_DIR:?}"/*
    mkdir -p "${CRAWL_LEDGER_DATA_DIR}" "${SCRAPE_BLOB_DIR}"
  else
    "${VEIL_ROOT}/scripts/ops/compose-down-ephemeral.sh"
  fi
fi

if [[ "$FULL" -eq 1 ]]; then
  source_profile full-rebuild
else
  source_profile fast-rich
fi

log "starting full stack (sources=${SCRAPE_SOURCES}, NVD_MAX_PAGES=${NVD_MAX_PAGES}, FORCE_REFETCH=${SCRAPE_FORCE_REFETCH:-0})..."
exec "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
