#!/usr/bin/env bash
# Smoke-test MCP stdio server against a running Neo4j (compose graph stack).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=scripts/lib/common.sh
source "${ROOT}/scripts/lib/common.sh"

NEO4J_URI="${NEO4J_URI:-neo4j://localhost:7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
NEO4J_DB="${NEO4J_DB:-neo4j}"

RUNNER="${ROOT}/scripts/mcp/run-veil-mcp.sh"
MCP_BIN="${MCP_BIN:-}"
if [[ -z "${MCP_BIN}" ]]; then
  if [[ -x "${RUNNER}" ]]; then
    MCP_BIN="${RUNNER}"
  else
    MCP_BIN="${ROOT}/knowledge/serve/bin/mcp"
    if [[ ! -x "${MCP_BIN}" ]]; then
      echo "Building mcp binary..." >&2
      (cd "${ROOT}/knowledge/serve" && env GOWORK="${ROOT}/knowledge/go.work" go build -o bin/mcp ./cmd/mcp)
    fi
  fi
fi

send_rpc() {
  local id="$1" method="$2" params="${3:-null}"
  local body
  body=$(printf '{"jsonrpc":"2.0","id":%s,"method":"%s","params":%s}' "$id" "$method" "$params")
  printf 'Content-Length: %d\r\n\r\n%s' "${#body}" "$body"
}

echo "MCP smoke: initialize (2024-11-05 + 2025-03-26) + tools/list + ti_health (Neo4j ${NEO4J_URI})"

run_smoke() {
  local init_ver="$1"
  {
    send_rpc 1 initialize "{\"protocolVersion\":\"${init_ver}\",\"capabilities\":{}}"
    send_rpc 2 ping '{}'
    send_rpc 3 "tools/list" '{}'
    send_rpc 4 "tools/call" '{"name":"ti_health","arguments":{}}'
  } | NEO4J_URI="$NEO4J_URI" NEO4J_USER="$NEO4J_USER" NEO4J_PASS="$NEO4J_PASS" NEO4J_DB="$NEO4J_DB" \
    timeout 30 "${MCP_BIN}"
}

out=$(run_smoke "2024-11-05" 2>&1) || true
if ! grep -q '"name":"veil-mcp"' <<<"$out" && ! grep -q '"name": "veil-mcp"' <<<"$out"; then
  out=$(run_smoke "2025-03-26" 2>&1) || true
fi

if ! grep -q '"name":"veil-mcp"' <<<"$out" && ! grep -q '"name": "veil-mcp"' <<<"$out"; then
  echo "FAIL: initialize response missing veil-mcp" >&2
  echo "$out" | head -c 2000 >&2
  exit 1
fi

if ! grep -q 'ti_list_categories' <<<"$out"; then
  echo "FAIL: tools/list missing ti_list_categories" >&2
  exit 1
fi

if ! grep -q '"neo4j_ok": true' <<<"$out" && ! grep -q '"neo4j_ok":true' <<<"$out"; then
  echo "WARN: ti_health did not report neo4j_ok (is Neo4j up and bootstrapped?)" >&2
  echo "$out" | tail -c 1500 >&2
  exit 1
fi

echo "OK: MCP smoke passed"
