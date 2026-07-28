#!/usr/bin/env bash
# Import a graph pack ZIP: verify manifest sha256 and schema, stream Cypher to Neo4j.
# Usage: ./scripts/graph-pack/import.sh <https://.../veil-graph-vX.zip | /path/to/*.zip>
# Env: NEO4J_*, USE_DOCKER_COMPOSE=1
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

SOURCE="${1:?usage: $0 <https://.../*.zip | /path/to/*.zip>}"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

if [[ "${SOURCE}" == https://* ]] || [[ "${SOURCE}" == http://* ]]; then
  echo "Downloading pack..."
  curl -fsSL "${SOURCE}" -o "${TMP}/pack.zip"
else
  cp "${SOURCE}" "${TMP}/pack.zip"
fi

unzip -q "${TMP}/pack.zip" -d "${TMP}/ex"
M="${TMP}/ex/manifest.json"
CY="${TMP}/ex/graph.cypher"
if [[ ! -f "${M}" ]] || [[ ! -f "${CY}" ]]; then
  echo "ZIP must contain manifest.json and graph.cypher at archive root" >&2
  exit 1
fi

GV="$(validate_pack_manifest "${M}")"
EXP="$(python3 -c "import json; print(json.load(open('${M}'))['sha256'])")"
GOT="$(sha256_file "${CY}")"
if [[ "${EXP}" != "${GOT}" ]]; then
  echo "sha256 mismatch: manifest has ${EXP}, file has ${GOT}" >&2
  exit 1
fi
echo "Checksum OK (graphVersion=${GV}, schema=$(python3 -c "import json; print(json.load(open('${M}'))['schema'])"))"

run_shell() {
  if [[ "${USE_DOCKER_COMPOSE:-}" == "1" ]]; then
    compose exec -T neo4j cypher-shell \
      -a "${NEO4J_URI}" -u "${NEO4J_USER}" -p "${NEO4J_PASS}" -d "${NEO4J_DB}" "$@"
  elif command -v cypher-shell >/dev/null 2>&1; then
    cypher-shell -a "${NEO4J_URI}" -u "${NEO4J_USER}" -p "${NEO4J_PASS}" -d "${NEO4J_DB}" "$@"
  else
    echo "cypher-shell not in PATH; set USE_DOCKER_COMPOSE=1 from repo root" >&2
    exit 1
  fi
}

echo "Streaming Cypher (this may take a while)..."
run_shell < "${CY}"
echo "Import finished."
