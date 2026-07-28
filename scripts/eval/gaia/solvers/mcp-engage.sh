#!/usr/bin/env bash
# Optional GAIA solver via veil-engage MCP (not used in default CI).
# Requires: running engage MCP, model credentials, GAIA attachments resolved under GAIA_DATA_DIR.
set -euo pipefail
TASKS="${1:?metadata.jsonl}"
MCP_URL="${VEIL_ENGAGE_MCP_URL:-http://127.0.0.1:8891/mcp}"

if [[ "${GAIA_SOLVER_DRY_RUN:-}" == "1" ]]; then
  echo "[mcp-engage] DRY_RUN: would call ${MCP_URL} per task" >&2
  exec "$(dirname "$0")/stub.sh" "${TASKS}"
fi

echo "[mcp-engage] not implemented in-repo; wire your agent client to ${MCP_URL}" >&2
echo "Use GAIA_SOLVER=scripts/eval/gaia/solvers/stub.sh for pipeline smoke." >&2
exit 2
