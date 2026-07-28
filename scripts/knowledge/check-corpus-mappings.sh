#!/usr/bin/env bash
# Verify committed playbook framework mappings (V0 SOT).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MAP="${ROOT}/pkg/playbook/corpus/mappings"
required=(
  "$MAP/README.md"
  "$MAP/attack-navigator-layer.json"
  "$MAP/mitre-attack/coverage-summary.md"
  "$MAP/nist-csf/README.md"
  "$MAP/owasp/README.md"
)
for f in "${required[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "missing: $f" >&2
    exit 1
  fi
done
python3 -c "
import json, sys
p = sys.argv[1]
with open(p) as f:
    d = json.load(f)
assert 'techniques' in d or 'name' in d, 'unexpected navigator layer shape'
print('OK: navigator layer valid JSON')
" "$MAP/attack-navigator-layer.json"
echo "OK: corpus mappings (${#required[@]} required files)"
