#!/usr/bin/env bash
set -euo pipefail

NEO4J_URI="${NEO4J_URI:-neo4j://neo4j:7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
NEO4J_DB="${NEO4J_DB:-neo4j}"

GRAPH_PACK_DEFAULT_VERSION="${GRAPH_PACK_DEFAULT_VERSION:-v0.4.0}"
DEFAULT_PACK_URL="${GRAPH_PACK_DEFAULT_URL:-https://github.com/butbeautifulv/veil/releases/download/veil-graph-${GRAPH_PACK_DEFAULT_VERSION}/veil-graph-${GRAPH_PACK_DEFAULT_VERSION}.zip}"

shell() {
  cypher-shell -a "${NEO4J_URI}" -u "${NEO4J_USER}" -p "${NEO4J_PASS}" -d "${NEO4J_DB}" "$@"
}

echo "graph-bootstrap: waiting for Neo4j..."
for _ in $(seq 1 90); do
  if shell 'RETURN 1' >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
if ! shell 'RETURN 1' >/dev/null 2>&1; then
  echo "Neo4j not reachable" >&2
  exit 1
fi

if [[ "${GRAPH_PACK_SKIP:-0}" == "1" ]]; then
  echo "GRAPH_PACK_SKIP=1 — skipping import"
  exit 0
fi

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

resolve_zip() {
  local rel="/pack/releases/veil-graph-${GRAPH_PACK_DEFAULT_VERSION}.zip"
  if [[ -f "${rel}" ]] && [[ -s "${rel}" ]]; then
    echo "Using local pack ${rel}"
    cp "${rel}" "${TMP}/pack.zip"
    return
  fi
  if [[ -f /pack/host.zip ]] && [[ -s /pack/host.zip ]]; then
    echo "Using bind-mounted /pack/host.zip"
    cp /pack/host.zip "${TMP}/pack.zip"
    return
  fi
  if [[ -n "${GRAPH_PACK_FILE:-}" ]] && [[ -f "${GRAPH_PACK_FILE}" ]]; then
    echo "Using local GRAPH_PACK_FILE=${GRAPH_PACK_FILE}"
    cp "${GRAPH_PACK_FILE}" "${TMP}/pack.zip"
    return
  fi
  if [[ -n "${GRAPH_PACK_URL:-}" ]]; then
    echo "Downloading GRAPH_PACK_URL..."
    curl -fsSL "${GRAPH_PACK_URL}" -o "${TMP}/pack.zip"
    return
  fi
  if [[ "${GRAPH_PACK_DEFAULT:-1}" == "1" ]]; then
    echo "Downloading default pack from ${DEFAULT_PACK_URL}"
    curl -fsSL "${DEFAULT_PACK_URL}" -o "${TMP}/pack.zip"
    return
  fi
  echo "No GRAPH_PACK_FILE, GRAPH_PACK_URL, or GRAPH_PACK_DEFAULT=1 — nothing to import"
  exit 0
}

resolve_zip

unzip -q "${TMP}/pack.zip" -d "${TMP}/ex"
M="${TMP}/ex/manifest.json"
CY="${TMP}/ex/graph.cypher"
if [[ ! -f "${M}" ]] || [[ ! -f "${CY}" ]]; then
  echo "ZIP must contain manifest.json and graph.cypher at archive root" >&2
  exit 1
fi

EXP="$(python3 -c "import json; print(json.load(open('${M}'))['sha256'])")"
GOT="$(python3 -c "import hashlib; print(hashlib.sha256(open('${CY}','rb').read()).hexdigest())")"
if [[ "${EXP}" != "${GOT}" ]]; then
  echo "sha256 mismatch: manifest=${EXP} file=${GOT}" >&2
  exit 1
fi
echo "Checksum OK; streaming Cypher into Neo4j (may take several minutes)..."
shell < "${CY}"
echo "graph-bootstrap: import finished."
