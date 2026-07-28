#!/usr/bin/env bash
# Build graph pack ZIP and create GitHub release veil-graph-vX.Y.Z.
# Usage: ./scripts/release/publish-graph-pack.sh [--draft] [--skip-build]
# Env: GRAPH_PACK_VERSION (default from versions.env), BUILD_PROFILE
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

DRAFT=""
SKIP_BUILD=0
for arg in "$@"; do
  case "$arg" in
    --draft) DRAFT="--draft" ;;
    --skip-build) SKIP_BUILD=1 ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

VERSION="$(pack_normalize_version "${GRAPH_PACK_VERSION}")"
TAG="$(pack_release_tag "${VERSION}")"
ZIP="${PACK_RELEASES_DIR}/$(pack_zip_name "${VERSION}")"
TEMPLATE="${VEIL_ROOT}/docs/templates/graph-pack-release-notes.md"
NOTES_FILE="$(mktemp)"
trap 'rm -f "${NOTES_FILE}"' EXIT

if [[ "$SKIP_BUILD" -eq 0 ]]; then
  EXPORT_FIRST=1 "${VEIL_ROOT}/scripts/graph-pack/build.sh" "${VERSION}"
fi

if [[ ! -f "${ZIP}" ]]; then
  echo "missing pack: ${ZIP}" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI required for publish" >&2
  exit 1
fi

PREV_TAG="$(git tag -l 'veil-graph-v*' --sort=-v:refname 2>/dev/null | head -1 || true)"
INGEST_PATHS=(
  discovery/harvest/internal/sources
  pipeline/ned/internal/sources
  knowledge/ingest/internal/sources
  pkg/harvest
  pkg/commit
  docs/schemas
)
if [[ -n "${PREV_TAG}" ]]; then
  INGEST_CHANGELOG="$(git log --oneline "${PREV_TAG}..HEAD" -- "${INGEST_PATHS[@]}" 2>/dev/null | sed 's/^/- /' || true)"
else
  INGEST_CHANGELOG="- (no previous veil-graph-v* tag)"
fi
[[ -z "${INGEST_CHANGELOG}" ]] && INGEST_CHANGELOG="- (no ingest-path commits since ${PREV_TAG:-initial})"

export GRAPH_PACK_VERSION="${VERSION}"
export BUILD_PROFILE="${BUILD_PROFILE:-fast-rich}"
export INGEST_CHANGELOG
export NODE_COUNTS="${NODE_COUNTS:-}"

if [[ -f "${TEMPLATE}" ]]; then
  envsubst '${GRAPH_PACK_VERSION} ${BUILD_PROFILE} ${INGEST_CHANGELOG} ${NODE_COUNTS}' \
    <"${TEMPLATE}" >"${NOTES_FILE}"
else
  echo "Graph pack ${VERSION} for Veil (Neo4j 5.x)." >"${NOTES_FILE}"
fi

TITLE="${TAG} — Veil graph pack"
gh_args=(release create "${TAG}" --repo butbeautifulv/veil --title "${TITLE}" --notes-file "${NOTES_FILE}")
[[ -n "${DRAFT}" ]] && gh_args+=(--draft)
gh "${gh_args[@]}" "${ZIP}"

echo "Published: ${TAG}"
echo "GRAPH_PACK_DEFAULT_URL=$(pack_release_url "${VERSION}")"
