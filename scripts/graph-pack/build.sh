#!/usr/bin/env bash
# Build a versioned graph pack ZIP (manifest.json + graph.cypher) for GitHub Releases.
# Usage: ./scripts/graph-pack/build.sh [v0.4.0]
# Env: EXPORT_FIRST=1, GRAPH_PACK_VERSION
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

VERSION="${GRAPH_PACK_VERSION:-${1:-}}"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(date -u +%Y%m%dT%H%MZ)"
fi
VERSION="$(pack_normalize_version "${VERSION}")"

mkdir -p "${PACK_RELEASES_DIR}"

if [[ "${EXPORT_FIRST:-}" == "1" ]]; then
  "$(dirname "$0")/export-cypher.sh"
fi

if [[ ! -f "${NEO4J_EXPORT_CYPHER}" ]]; then
  echo "Missing ${NEO4J_EXPORT_CYPHER}. Run export-cypher.sh or EXPORT_FIRST=1 ./scripts/graph-pack/build.sh" >&2
  exit 1
fi

SHA="$(sha256_file "${NEO4J_EXPORT_CYPHER}")"
CREATED="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

cp "${NEO4J_EXPORT_CYPHER}" "${TMP}/graph.cypher"
export VERSION CREATED SHA PACK_SCHEMA
python3 - <<'PY' > "${TMP}/manifest.json"
import json, os
m = {
    "schema": os.environ["PACK_SCHEMA"],
    "graphVersion": os.environ["VERSION"],
    "createdAt": os.environ["CREATED"],
    "cypherFile": "graph.cypher",
    "sha256": os.environ["SHA"],
    "neo4j": "5.x",
    "notes": "Import with scripts/graph-pack/import.sh. Plain Cypher from APOC apoc.export.cypher.all; target Neo4j 5.x.",
}
print(json.dumps(m, indent=2))
PY

ZIP="${PACK_RELEASES_DIR}/$(pack_zip_name "${VERSION}")"
( cd "${TMP}" && zip -q "${ZIP}" manifest.json graph.cypher )
cp "${TMP}/manifest.json" "${PACK_RELEASES_DIR}/manifest.${VERSION}.json"
echo "Pack: ${ZIP}"
echo "Manifest: ${PACK_RELEASES_DIR}/manifest.${VERSION}.json"
echo "sha256(graph.cypher)= ${SHA}"
echo "Release tag: $(pack_release_tag "${VERSION}")"
echo "Release URL: $(pack_release_url "${VERSION}")"
