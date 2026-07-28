#!/usr/bin/env bash
# Export Neo4j graph to var/veil/graph/graph.cypher (requires APOC).
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

mkdir -p "${NEO4J_EXPORT_HOST_DIR}"
OUT_REL="user_export/graph.cypher"

if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
  echo "docker compose not available; run manually against Neo4j:" >&2
  echo "  cypher-shell ... \"CALL apoc.export.cypher.all('${OUT_REL}', {format: 'cypher-shell'})\"" >&2
  exit 1
fi

compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" -d neo4j \
  "CALL apoc.export.cypher.all('${OUT_REL}', {format: 'cypher-shell'}) YIELD file, batches, source, format RETURN file, batches, source, format;"

echo "Export written to ${NEO4J_EXPORT_CYPHER}"
