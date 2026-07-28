#!/usr/bin/env bash
# Bump GRAPH_PACK_VERSION in versions.env and propagate to docs / compose.
# Usage: ./scripts/release/bump-graph-version.sh [patch|minor]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

BUMP="${1:-patch}"
VERSIONS_FILE="${VEIL_ROOT}/versions.env"

# Files that embed the default graph pack version string (updated on bump).
DOC_FILES=(
  "${VEIL_ROOT}/README.md"
  "${VEIL_ROOT}/deploy/README.md"
  "${VEIL_ROOT}/docs/architecture/threatintel-runtime.md"
  "${VEIL_ROOT}/docs/contracts/graph-pack.md"
  "${VEIL_ROOT}/scripts/README.md"
)
TESTPACK_YML="${VEIL_ROOT}/docker-compose.testpack.yml"

read_versions() {
  # shellcheck disable=SC1090
  source "${VERSIONS_FILE}"
}

write_versions() {
  cat >"${VERSIONS_FILE}" <<EOF
# Single source of truth for app and graph-pack versions (sourced by scripts/lib/common.sh).
APP_VERSION=${APP_VERSION}
GRAPH_PACK_VERSION=${GRAPH_PACK_VERSION}
EOF
}

bump_semver() {
  local v="${1#v}" kind="$2"
  local major minor patch
  IFS=. read -r major minor patch <<<"${v}"
  patch="${patch:-0}"
  minor="${minor:-0}"
  major="${major:-0}"
  case "${kind}" in
    patch)
      patch=$((patch + 1))
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    *)
      echo "usage: $0 [patch|minor]" >&2
      exit 1
      ;;
  esac
  echo "v${major}.${minor}.${patch}"
}

read_versions
OLD_GRAPH="${GRAPH_PACK_VERSION}"
OLD_APP="${APP_VERSION}"
NEW_GRAPH="$(bump_semver "${OLD_GRAPH}" "${BUMP}")"
NEW_APP="${NEW_GRAPH#v}"

GRAPH_PACK_VERSION="${NEW_GRAPH}"
APP_VERSION="${NEW_APP}"
write_versions
echo "${APP_VERSION}" >"${VEIL_ROOT}/VERSION"

OLD_ZIP="$(pack_zip_name "${OLD_GRAPH}")"
NEW_ZIP="$(pack_zip_name "${NEW_GRAPH}")"
OLD_TAG="$(pack_release_tag "${OLD_GRAPH}")"
NEW_TAG="$(pack_release_tag "${NEW_GRAPH}")"

for f in "${DOC_FILES[@]}"; do
  [[ -f "${f}" ]] || continue
  sed -i \
    -e "s|${OLD_GRAPH}|${NEW_GRAPH}|g" \
    -e "s|${OLD_TAG}|${NEW_TAG}|g" \
    -e "s|${OLD_ZIP}|${NEW_ZIP}|g" \
    "${f}"
done

if [[ -f "${TESTPACK_YML}" ]]; then
  sed -i \
    -e "s|${OLD_ZIP}|${NEW_ZIP}|g" \
    "${TESTPACK_YML}"
fi

echo "Bumped GRAPH_PACK_VERSION: ${OLD_GRAPH} -> ${NEW_GRAPH}"
echo "APP_VERSION / VERSION: ${NEW_APP}"
echo ""
echo "Next: rebuild and publish the pack if graph data changed:"
echo "  ./scripts/graph-pack/profile-fast-rich.sh   # ~25 min"
echo "  GRAPH_PACK_VERSION=${NEW_GRAPH} ./scripts/graph-pack/build.sh"
echo "  GRAPH_PACK_VERSION=${NEW_GRAPH} ./scripts/release/publish-graph-pack.sh --skip-build"
