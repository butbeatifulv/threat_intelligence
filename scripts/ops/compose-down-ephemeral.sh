#!/usr/bin/env bash
# Stop stack and remove only ephemeral volumes (neo4j_data, nats_data).
# Preserves host bind mounts: var/veil/blobs, var/veil/ledger/mysql, var/veil/graph.
# Usage: ./scripts/ops/compose-down-ephemeral.sh
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[compose-down-ephemeral] %s\n' "$*"; }

log "stopping services (keeping var/veil crawl state on host)..."
compose_with_extras down --remove-orphans 2>/dev/null || true

log "removing neo4j_data and nats_data volumes if present..."
while IFS= read -r vol; do
  [[ -z "${vol}" ]] && continue
  docker volume rm "${vol}" 2>/dev/null || true
done < <(docker volume ls -q | grep -E 'neo4j_data$|nats_data$' || true)

log "done — crawl ledger (var/veil/ledger) and blobs (var/veil/blobs) retained"
