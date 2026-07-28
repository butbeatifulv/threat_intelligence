#!/usr/bin/env bash
# Classify scripts/ files as KEEP (documented), MAKE, CI, or ORPHAN (no Makefile/workflow ref).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

refs_file="$(mktemp)"
trap 'rm -f "$refs_file"' EXIT

{
  rg -l 'scripts/' Makefile .github/workflows 2>/dev/null || true
  rg -o 'scripts/[a-zA-Z0-9_./-]+' Makefile .github/workflows 2>/dev/null | sort -u || true
  rg -l 'scripts/' docs/*.md scripts/README.md AGENTS.md CONTRIBUTING.md README.md 2>/dev/null || true
} >"$refs_file" || true

echo "# Script classification (generated $(date -u +%Y-%m-%dT%H:%MZ))"
echo ""
echo "| Script | Class |"
echo "|--------|-------|"

while IFS= read -r -d '' f; do
  rel="${f#"$ROOT"/}"
  class=ORPHAN
  if grep -qF "$rel" "$refs_file" 2>/dev/null || grep -qF "${rel#scripts/}" "$refs_file" 2>/dev/null; then
    if grep -qF "$rel" Makefile 2>/dev/null; then
      class=MAKE
    elif grep -qF "$rel" .github/workflows/*.yml 2>/dev/null; then
      class=CI
    else
      class=KEEP
    fi
  fi
  # README-listed scripts
  if [[ "$class" == ORPHAN ]] && grep -qF "${rel#scripts/}" scripts/README.md 2>/dev/null; then
    class=KEEP
  fi
  echo "| \`$rel\` | $class |"
done < <(find scripts -type f \( -name '*.sh' -o -name '*.py' \) ! -path 'scripts/lib/*' ! -path 'scripts/test/lib/*' -print0 | sort -z)
