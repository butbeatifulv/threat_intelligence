#!/usr/bin/env bash
# Neo4j maintenance: duplicate rels, optional stale isolated IOCs.
# Usage: ./scripts/housekeeping/graph-dedup-cleanup.sh [--dry-run]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    -h|--help)
      sed -n '1,20p' "$0"
      exit 0
      ;;
    *)
      echo "unknown arg: $1" >&2
      exit 1
      ;;
  esac
  shift
done

STALE_DAYS="${GRAPH_IOC_STALE_DAYS:-90}"
DELETE_ISOLATED="${GRAPH_DELETE_STALE_ISOLATED_IOCS:-0}"

cutoff_iso() {
  python3 -c "import datetime; print((datetime.datetime.now(datetime.timezone.utc)-datetime.timedelta(days=int('${STALE_DAYS}'))).strftime('%Y-%m-%dT%H:%M:%S.000000000+00:00'))"
}
CUTOFF="$(cutoff_iso)"

run_cypher() {
  local q="$1"
  if command -v cypher-shell >/dev/null 2>&1; then
    cypher-shell --non-interactive -a "$NEO4J_URI" -u "$NEO4J_USER" -p "$NEO4J_PASS" -d "$NEO4J_DB" "$q"
  else
    echo "cypher-shell not found; install Neo4j tools or run inside neo4j container:" >&2
    exit 1
  fi
}

echo "== Duplicate HAS_ADVISORY (same Vulnerability -> SecurityAdvisory) =="
Q_DUP_COUNT='MATCH (v:Vulnerability)-[r:HAS_ADVISORY]->(a:SecurityAdvisory)
WITH v, a, collect(r) AS rels WHERE size(rels) > 1 RETURN count(*) AS dupGroups'
run_cypher "$Q_DUP_COUNT"

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "(dry-run) skipping DELETE of extra parallel HAS_ADVISORY relationships."
else
  echo "== Removing extra HAS_ADVISORY (keeping one per pair) =="
  Q_DUP_FIX='MATCH (v:Vulnerability)-[r:HAS_ADVISORY]->(a:SecurityAdvisory)
WITH v, a, collect(r) AS rels WHERE size(rels) > 1
FOREACH (x IN tail(rels) | DELETE x)'
  run_cypher "$Q_DUP_FIX"
fi

echo "== Isolated IOCs (no relationships), stale vs cutoff $CUTOFF =="
Q_ISO_COUNT="MATCH (i:IOC)
WHERE NOT (i)--()
WITH i, coalesce(i.lastSeen, i.updatedAt, '1970-01-01T00:00:00.000000000+00:00') AS ts
WHERE ts < '$CUTOFF'
RETURN count(i) AS staleIsolatedIOCs"
run_cypher "$Q_ISO_COUNT"

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "(dry-run) not deleting stale isolated IOCs."
  exit 0
fi

if [[ "$DELETE_ISOLATED" == "1" ]]; then
  echo "== DETACH DELETE stale isolated IOCs =="
  Q_ISO_DEL="MATCH (i:IOC)
WHERE NOT (i)--()
WITH i, coalesce(i.lastSeen, i.updatedAt, '1970-01-01T00:00:00.000000000+00:00') AS ts
WHERE ts < '$CUTOFF'
DETACH DELETE i"
  run_cypher "$Q_ISO_DEL"
else
  echo "Set GRAPH_DELETE_STALE_ISOLATED_IOCS=1 to remove stale isolated IOCs."
fi

echo "Done."
