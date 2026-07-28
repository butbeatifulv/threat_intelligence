#!/usr/bin/env bash
# Pilot-only solver: echoes golden answers (validates scoring pipeline, not agent quality).
set -euo pipefail
TASKS="${1:?metadata.jsonl}"
while IFS= read -r line; do
  [[ -z "${line}" ]] && continue
  python3 -c 'import json,sys; r=json.loads(sys.argv[1]); print(json.dumps({"task_id":r["task_id"],"prediction":r["Final answer"]}))' "${line}"
done < "${TASKS}"
