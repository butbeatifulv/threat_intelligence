#!/usr/bin/env bash
# Import Anthropic Cybersecurity Skills from .external into committed Veil paths.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${CORPUS_SRC:-$ROOT/.external/Anthropic-Cybersecurity-Skills-main}"
MAP_DST="$ROOT/pkg/playbook/corpus/mappings"
SKILLS_DST="$ROOT/corpus/anthropic-cybersecurity-skills/skills"
VERSION_FILE="$ROOT/pkg/playbook/corpus/VERSION"

IMPORT_MAPPINGS=1
IMPORT_SKILLS=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --mappings-only) IMPORT_SKILLS=0 ;;
    --skills-only) IMPORT_MAPPINGS=0 ;;
    --src) shift; SRC="${1:?}" ;;
    -h|--help)
      echo "Usage: corpus-import.sh [--mappings-only|--skills-only] [--src PATH]"
      exit 0
      ;;
  esac
  shift
done

if [[ ! -d "$SRC" ]]; then
  echo "ERROR: upstream not found: $SRC" >&2
  echo "Clone into .external/ or set CORPUS_SRC" >&2
  exit 1
fi

if [[ "$IMPORT_MAPPINGS" == 1 ]]; then
  [[ -d "$SRC/mappings" ]] || { echo "missing $SRC/mappings" >&2; exit 1; }
  mkdir -p "$MAP_DST/mitre-attack" "$MAP_DST/nist-csf" "$MAP_DST/owasp"
  rsync -a --delete "$SRC/mappings/" "$MAP_DST/"
  echo "Imported mappings -> $MAP_DST"
fi

if [[ "$IMPORT_SKILLS" == 1 ]]; then
  [[ -d "$SRC/skills" ]] || { echo "missing $SRC/skills" >&2; exit 1; }
  mkdir -p "$(dirname "$SKILLS_DST")"
  rsync -a --delete "$SRC/skills/" "$SKILLS_DST/"
  n="$(find "$SKILLS_DST" -name SKILL.md | wc -l)"
  echo "Imported skills ($n SKILL.md) -> $SKILLS_DST"
fi

sha="unknown"
if [[ -d "$SRC/.git" ]]; then
  sha="$(git -C "$SRC" rev-parse HEAD 2>/dev/null || echo unknown)"
fi
cat > "$VERSION_FILE" <<EOF
upstream_repo=https://github.com/anthropics/anthropic-cybersecurity-skills
upstream_path=Anthropic-Cybersecurity-Skills
imported_at=$(date -u +%Y-%m-%d)
upstream_sha=${sha}
skills_count=$(find "$SKILLS_DST" -name SKILL.md 2>/dev/null | wc -l | tr -d ' ')
mappings_files=$(find "$MAP_DST" -type f 2>/dev/null | wc -l | tr -d ' ')
EOF
echo "Wrote $VERSION_FILE (sha=$sha)"
