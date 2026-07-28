#!/usr/bin/env bash
# List agent and phase ids from manifest.yaml
set -euo pipefail
MANIFEST="$(cd "$(dirname "$0")/../.." && pwd)/.cursor/agents/manifest.yaml"
python3 - "$MANIFEST" <<'PY'
import sys, yaml
path = sys.argv[1]
with open(path, encoding="utf-8") as f:
    doc = yaml.safe_load(f)
print("agents:")
for a in doc.get("agents", []):
    print(f"  - {a['id']}: {a.get('description', '')}")
print("phases:")
for p in doc.get("phases", []):
    print(f"  - {p['id']}: agent={p.get('agent')} branch={p.get('branch', '')}")
PY
