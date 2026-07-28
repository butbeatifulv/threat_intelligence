#!/usr/bin/env bash
# Clone openJiuwen agent-store into .external/ (reference only; directory is gitignored).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DEST="${ROOT}/.external/agent-store"
URL="${AGENT_STORE_URL:-https://gitcode.com/openJiuwen/agent-store.git}"

mkdir -p "${ROOT}/.external"
if [[ -d "${DEST}/.git" ]]; then
  echo "[clone-agent-store] already present at ${DEST}; fetch latest"
  git -C "${DEST}" fetch --depth 1 origin
  git -C "${DEST}" checkout -f FETCH_HEAD 2>/dev/null || git -C "${DEST}" pull --ff-only
else
  git clone --depth 1 "${URL}" "${DEST}"
fi
echo "[clone-agent-store] ok: ${DEST}"
echo "See docs/agents/external-agent-store.md — reference only, do not run community agents in CI/prod."
