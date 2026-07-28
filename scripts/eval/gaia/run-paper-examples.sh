#!/usr/bin/env bash
# GAIA paper public examples (arXiv:2311.12983 Fig. 1) — scorer/format harness only.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
TASKS="${GAIA_TASKS:-${ROOT}/eval/gaia/fixtures/paper-examples/metadata.jsonl}"
SOLVER="${GAIA_SOLVER:-${ROOT}/scripts/eval/gaia/solvers/stub.sh}"
OUT_DIR="${ROOT}/eval/gaia/results"
PRED="${OUT_DIR}/paper-predictions.jsonl"
METRICS="${OUT_DIR}/paper-metrics.json"

mkdir -p "${OUT_DIR}"
chmod +x "${SOLVER}" 2>/dev/null || true
"${SOLVER}" "${TASKS}" > "${PRED}"
python3 "${ROOT}/scripts/eval/gaia/score.py" --tasks "${TASKS}" --predictions "${PRED}" --out "${METRICS}"
echo "[gaia-paper] wrote ${METRICS} (stub checks format pipeline; real scores need live agent + web)"
