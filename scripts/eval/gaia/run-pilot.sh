#!/usr/bin/env bash
# GAIA pilot eval: offline fixtures only (no Hugging Face download).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
TASKS="${GAIA_TASKS:-${ROOT}/eval/gaia/fixtures/pilot/metadata.jsonl}"
SOLVER="${GAIA_SOLVER:-${ROOT}/scripts/eval/gaia/solvers/stub.sh}"
OUT_DIR="${ROOT}/eval/gaia/results"
PRED="${OUT_DIR}/pilot-predictions.jsonl"
METRICS="${OUT_DIR}/pilot-metrics.json"

mkdir -p "${OUT_DIR}"
chmod +x "${SOLVER}" 2>/dev/null || true

if [[ ! -f "${TASKS}" ]]; then
  echo "[gaia-pilot] missing tasks: ${TASKS}" >&2
  exit 1
fi

echo "[gaia-pilot] tasks=${TASKS} solver=${SOLVER}"
"${SOLVER}" "${TASKS}" > "${PRED}"
python3 "${ROOT}/scripts/eval/gaia/score.py" --tasks "${TASKS}" --predictions "${PRED}" --out "${METRICS}"
echo "[gaia-pilot] wrote ${METRICS}"
