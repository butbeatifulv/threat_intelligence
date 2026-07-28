#!/usr/bin/env bash
# Unit/smoke helpers for shell tests — source from scripts/test/*.sh

unit_skip_no_go() {
  if ! command -v go >/dev/null 2>&1; then
    echo "[unit] SKIP: go not available" >&2
    exit 0
  fi
}

unit_assert_json_field() {
  local json=$1 field=$2 expected=$3
  local got
  got=$(printf '%s' "${json}" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d${field})" 2>/dev/null) || {
    echo "[unit] invalid json or field ${field}" >&2
    return 1
  }
  if [[ "${got}" != "${expected}" ]]; then
    echo "[unit] ${field}: want ${expected}, got ${got}" >&2
    return 1
  fi
}

unit_read_fixture() {
  local root=${VEIL_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)}
  local path=$1
  if [[ ! -f "${path}" ]]; then
    path="${root}/${path}"
  fi
  cat "${path}"
}
