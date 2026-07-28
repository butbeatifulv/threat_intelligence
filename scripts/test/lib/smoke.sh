#!/usr/bin/env bash
# Shared smoke helpers — source from scripts/test/smoke-*.sh

smoke_skip_no_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "[smoke] SKIP: docker not available" >&2
    exit 0
  fi
  if ! docker info >/dev/null 2>&1; then
    echo "[smoke] SKIP: docker daemon not running" >&2
    exit 0
  fi
}

# smoke_wait_http URL [MAX_SEC] [LABEL] [POLL_SEC]
# LABEL is included in timeout messages when set.
smoke_wait_http() {
  local url=$1 max=${2:-60} label=${3:-} step=${4:-1}
  local i=0
  while [[ $i -lt $max ]]; do
    if curl -sf "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$step"
    i=$((i + step))
  done
  if [[ -n "$label" ]]; then
    echo "[smoke] timeout waiting for ${label} (${url}, ${max}s)" >&2
  else
    echo "[smoke] timeout waiting for ${url} (${max}s)" >&2
  fi
  return 1
}
