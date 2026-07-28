#!/usr/bin/env bash
# Warn on markdown links that likely target directories but omit a trailing slash.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

cd "${VEIL_ROOT}"
errors=0

DIR_SUFFIXES=(
'discovery/' 'pipeline/' 'graph/' 'deploy/' 'pkg/' 'docs/' 'scripts/'
  'harvest/' 'connector/' 'ned/' 'ingest/' 'serve/' 'sources/'
  'internal/' 'domain/' 'feeds/' 'ledger/' 'schemas/'
)

while IFS= read -r -d '' f; do
  while IFS= read -r line; do
    [[ "${line}" == *"](http"* ]] && continue
    [[ "${line}" == *"](#"* ]] && continue
    [[ "${line}" != *"]("* ]] && continue
    href="${line#*](}"
    href="${href%)}"
    case "${href}" in
      */) continue ;;
      *.*) continue ;;  # likely a file
    esac
    for seg in "${DIR_SUFFIXES[@]}"; do
      case "${href}" in
        *"${seg}"|*"${seg%/}")
          echo "${f}: directory link without trailing slash: ${href}" >&2
          errors=$((errors + 1))
          break
          ;;
      esac
    done
  done < <(grep -E '\]\([^)]+\)' "${f}" 2>/dev/null || true)
done < <(find . -name '*.md' -not -path './data/*' -not -path './.cursor/plans/*' -print0 2>/dev/null)

if [[ "${errors}" -gt 0 ]]; then
  echo "${errors} issue(s); use path/to/dir/ for directory links (see docs/agents/coding-style.md)" >&2
  exit 1
fi
echo "markdown directory links OK"
