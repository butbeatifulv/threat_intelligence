# Veil shell helpers — source from scripts/* (do not execute directly).
# shellcheck shell=bash

if [[ -z "${VEIL_ROOT:-}" ]]; then
  _veil_lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  VEIL_ROOT="$(cd "${_veil_lib_dir}/../.." && pwd)"
fi

# shellcheck disable=SC1091
source "${VEIL_ROOT}/versions.env"

COMPOSE="${COMPOSE:-docker compose}"
VEIL_COMPOSE_FILES="-f deploy/discovery/compose.yml -f deploy/pipeline/compose.yml -f deploy/knowledge/compose.yml -f deploy/knowledge/compose.neo4j-publish.yml"
COMPOSE_FILES="${COMPOSE_FILES:-${VEIL_COMPOSE_FILES}}"

PACK_BASENAME="veil-graph"
PACK_SCHEMA="veil.graph-pack/1"
PACK_SCHEMA_LEGACY="threat-intelligence.graph-pack/1"
GRAPH_PACK_DEFAULT_VERSION="${GRAPH_PACK_DEFAULT_VERSION:-${GRAPH_PACK_VERSION}}"

# Persistent host state (crawl blobs, ledger, graph pack artifacts).
VEIL_VAR_DIR="${VEIL_VAR_DIR:-${VEIL_ROOT}/var/veil}"
SCRAPE_BLOB_DIR="${SCRAPE_BLOB_DIR:-${VEIL_VAR_DIR}/blobs}"
CRAWL_LEDGER_DATA_DIR="${CRAWL_LEDGER_DATA_DIR:-${VEIL_VAR_DIR}/ledger/mysql}"
GRAPH_PACK_DIR="${GRAPH_PACK_DIR:-${VEIL_VAR_DIR}/graph}"

NEO4J_EXPORT_HOST_DIR="${NEO4J_EXPORT_HOST_DIR:-${GRAPH_PACK_DIR}}"
NEO4J_EXPORT_CYPHER="${NEO4J_EXPORT_HOST_DIR}/graph.cypher"
PACK_RELEASES_DIR="${NEO4J_EXPORT_HOST_DIR}/releases"

NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
NEO4J_DB="${NEO4J_DB:-neo4j}"
NEO4J_URI="${NEO4J_URI:-neo4j://localhost:7687}"

# Normalize version: accept v0.4.0 or 0.4.0 → v0.4.0
pack_normalize_version() {
  local v="${1#v}"
  echo "v${v}"
}

pack_zip_name() {
  local v
  v="$(pack_normalize_version "$1")"
  echo "${PACK_BASENAME}-${v}.zip"
}

pack_release_tag() {
  pack_zip_name "$1" | sed 's/.zip$//'
}

pack_release_url() {
  local tag zip
  tag="$(pack_release_tag "$1")"
  zip="$(pack_zip_name "$1")"
  echo "https://github.com/butbeautifulv/veil/releases/download/${tag}/${zip}"
}

compose() {
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "$@")
}

compose_with_extras() {
  local -a extra=()
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    extra+=(-f deploy/compose.scale.yml)
    export COMPOSE_PROFILES="${COMPOSE_PROFILES:-scrape-partition}"
  fi
  # shellcheck disable=SC2086
  (cd "${VEIL_ROOT}" && ${COMPOSE} ${COMPOSE_FILES} "${extra[@]}" "$@")
}

sha256_file() {
  local f="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$f" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$f" | awk '{print $1}'
  else
    openssl dgst -sha256 "$f" | awk '{print $2}'
  fi
}

validate_pack_manifest() {
  local manifest="$1"
  python3 - <<'PY' "$manifest"
import json, sys
m = json.load(open(sys.argv[1]))
schema = m.get("schema", "")
allowed = {"veil.graph-pack/1", "threat-intelligence.graph-pack/1"}
if schema not in allowed:
    raise SystemExit(f"unsupported manifest schema: {schema!r} (want one of {allowed})")
for key in ("graphVersion", "createdAt", "cypherFile", "sha256", "neo4j"):
    if key not in m:
        raise SystemExit(f"manifest missing {key!r}")
if m["cypherFile"] != "graph.cypher":
    raise SystemExit("manifest cypherFile must be graph.cypher")
print(m["graphVersion"])
PY
}

source_profile() {
  local name="$1"
  local f="${VEIL_ROOT}/deploy/profiles/${name}.env"
  if [[ ! -f "${f}" ]]; then
    echo "profile not found: ${f}" >&2
    return 1
  fi
  set -a
  # shellcheck disable=SC1090
  source "${f}"
  set +a
}
