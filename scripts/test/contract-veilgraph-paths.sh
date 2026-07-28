#!/usr/bin/env bash
# Contract smoke: veil-api engage read paths (veneno client lives in veneno repo).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ROUTER="$ROOT/knowledge/serve/internal/transport/httpserver/router.go"

paths=(
  "/v1/categories/engage/context"
  "/v1/playbooks/"
  "/v1/categories/"
)

fail=0
for p in "${paths[@]}"; do
  if ! grep -qF "$p" "$ROUTER" 2>/dev/null; then
    echo "router missing: $p" >&2
    fail=1
  fi
done

[[ $fail -eq 0 ]] || exit 1
echo "veil-api engage read paths OK"
