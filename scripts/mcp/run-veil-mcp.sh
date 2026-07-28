#!/usr/bin/env bash
# Launch veil-mcp for MCP clients (stdio). Logs go to stderr; stdout is JSON-RPC only.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

MCP_BIN="${VEIL_MCP_BIN:-${VEIL_ROOT}/knowledge/serve/bin/mcp}"
if [[ ! -x "${MCP_BIN}" ]]; then
  echo "[veil-mcp] building ${MCP_BIN}..." >&2
  (cd "${VEIL_ROOT}/knowledge/serve" && env GOWORK="${VEIL_ROOT}/knowledge/go.work" go build -o bin/mcp ./cmd/mcp)
fi

export NEO4J_URI="${NEO4J_URI:-neo4j://localhost:7687}"
export NEO4J_USER="${NEO4J_USER:-neo4j}"
export NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
export NEO4J_DB="${NEO4J_DB:-neo4j}"
export MCP_ENV="${MCP_ENV:-local}"

exec "${MCP_BIN}" "$@"
