#!/usr/bin/env bash
# Render Task/subagent prompt from .cursor/agents/manifest.yaml
# Usage: render-task-prompt.sh <agent_id> [--phase <phase_id>] [--set key=val]
set -euo pipefail

VEIL_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MANIFEST="${VEIL_ROOT}/.cursor/agents/manifest.yaml"

agent_id="${1:-}"
shift || true
phase_id=""
set_args=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase)
      phase_id="${2:-}"
      shift 2
      ;;
    --set)
      set_args+=("${2:-}")
      shift 2
      ;;
    *)
      echo "unknown arg: $1" >&2
      exit 1
      ;;
  esac
done

[[ -n "${agent_id}" ]] || { echo "usage: $0 <agent_id> [--phase <id>] [--set k=v]" >&2; exit 1; }
command -v python3 >/dev/null || { echo "python3 required" >&2; exit 1; }

export MANIFEST agent_id phase_id
printf '%s\n' "${set_args[@]}" | python3 <<'PY'
import os, sys

try:
    import yaml
except ImportError:
    sys.stderr.write("Install PyYAML: pip install pyyaml\n")
    sys.exit(1)

manifest_path = os.environ["MANIFEST"]
agent_id = os.environ["agent_id"]
phase_id = os.environ.get("phase_id", "")

with open(manifest_path, encoding="utf-8") as f:
    doc = yaml.safe_load(f)

agents = {a["id"]: a for a in doc.get("agents", [])}
if agent_id not in agents:
    sys.stderr.write(f"unknown agent_id: {agent_id}\n")
    sys.exit(1)

ctx = dict(doc.get("defaults", {}))
ctx.update(agents[agent_id])

if phase_id:
    phases = {p["id"]: p for p in doc.get("phases", [])}
    if phase_id not in phases:
        sys.stderr.write(f"unknown phase_id: {phase_id}\n")
        sys.exit(1)
    ctx.update(phases[phase_id])

for line in sys.stdin:
    line = line.strip()
    if line and "=" in line:
        k, v = line.split("=", 1)
        ctx[k] = v

tpl = ctx.get("template", "")
if not tpl:
    sys.stderr.write("agent has no template\n")
    sys.exit(1)

class SafeDict(dict):
    def __missing__(self, key):
        return "{" + key + "}"

print(tpl.format_map(SafeDict(ctx)))

for key in ("subagent_type", "readonly", "branch", "plan_path"):
    if key in ctx:
        print(f"# {key}={ctx[key]}", file=sys.stderr)
PY
