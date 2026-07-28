#!/usr/bin/env bash
# Fail if ingest-affecting paths changed without a versions.env bump.
# Usage: ./scripts/release/check-graph-version-bump.sh [base-ref]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

BASE="${1:-main}"
INGEST_PATHS=(
  discovery/harvest/internal/sources
  pipeline/ned/internal/sources
  knowledge/ingest/internal/sources
  pkg/harvest
  pkg/commit
  docs/schemas
)

if ! git rev-parse --verify "${BASE}" >/dev/null 2>&1; then
  BASE="HEAD~1"
fi

changed=0
for p in "${INGEST_PATHS[@]}"; do
  if git diff --name-only "${BASE}"...HEAD -- "${p}" 2>/dev/null | grep -q .; then
    changed=1
    break
  fi
done
# Include unstaged/staged against BASE for pre-commit use
if [[ "${changed}" -eq 0 ]]; then
  for p in "${INGEST_PATHS[@]}"; do
    if git diff --name-only "${BASE}" -- "${p}" 2>/dev/null | grep -q .; then
      changed=1
      break
    fi
  done
fi

if [[ "${changed}" -eq 0 ]]; then
  exit 0
fi

if git diff --name-only "${BASE}"...HEAD -- versions.env 2>/dev/null | grep -q versions.env \
  || git diff --name-only "${BASE}" -- versions.env 2>/dev/null | grep -q versions.env; then
  exit 0
fi

echo "ingest-affecting files changed since ${BASE} but versions.env was not updated." >&2
echo "Run: ./scripts/release/bump-graph-version.sh patch" >&2
exit 1
