#!/usr/bin/env bash
# Export crawl_resource rows to JSON for backup before experiments.
# Usage: ./scripts/crawl/ledger-dump.sh [output.json]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

OUT="${1:-${VEIL_VAR_DIR}/ledger/crawl_resource-$(date -u +%Y%m%dT%H%M%SZ).json}"
mkdir -p "$(dirname "${OUT}")"

if ! compose ps crawl-db 2>/dev/null | grep -qE 'Up|running'; then
  echo "crawl-db not running — start stack first" >&2
  exit 1
fi

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT

compose exec -T crawl-db mysql -uveil -pveilpass veil_ledger -N -B -e \
  "SELECT resource_key, source, url, content_sha256, last_fetched_at, last_changed_at, fetch_policy FROM crawl_resource ORDER BY resource_key;" \
  > "${TMP}"

python3 - "${OUT}" "${TMP}" <<'PY'
import json, sys
out, path = sys.argv[1], sys.argv[2]
rows = []
with open(path, encoding="utf-8") as f:
    for line in f:
        line = line.rstrip("\n")
        if not line:
            continue
        parts = line.split("\t")
        if len(parts) < 7:
            continue
        rows.append({
            "resource_key": parts[0],
            "source": parts[1],
            "url": parts[2],
            "content_sha256": parts[3] or None,
            "last_fetched_at": parts[4],
            "last_changed_at": parts[5] or None,
            "fetch_policy": parts[6],
        })
with open(out, "w", encoding="utf-8") as f:
    json.dump({"schema": "veil.crawl-ledger/1", "rows": rows}, f, indent=2)
    f.write("\n")
print(out)
print(f"{len(rows)} rows")
PY
