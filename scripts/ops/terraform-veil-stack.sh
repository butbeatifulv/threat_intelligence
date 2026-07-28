#!/usr/bin/env bash
# Apply or destroy Veil stack from Terraform-generated env (see deploy/terraform/).
# Usage: terraform-veil-stack.sh up|down
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

ACTION="${1:-}"
ENV_FILE="${TERRAFORM_COMPOSE_ENV:-}"
if [[ -z "${ENV_FILE}" ]]; then
  echo "TERRAFORM_COMPOSE_ENV required (path to generated veil-compose.env)" >&2
  exit 1
fi
[[ -f "${ENV_FILE}" ]] || { echo "env file not found: ${ENV_FILE}" >&2; exit 1; }

ENABLE_ENGAGE="${TF_ENABLE_ENGAGE:-1}"
ENABLE_ENGAGE_EVENTS="${TF_ENABLE_ENGAGE_EVENTS:-1}"
COMPOSE_BUILD="${TF_COMPOSE_BUILD:-1}"

COMPOSE_FILES="${VEIL_COMPOSE_FILES}"
if [[ "${ENABLE_ENGAGE}" == "1" ]]; then
  COMPOSE_FILES="${COMPOSE_FILES} -f deploy/engage/compose.yml -f deploy/engage/compose.veil-stack.yml"
fi
export COMPOSE_FILES

set -a
# shellcheck disable=SC1090
source "${ENV_FILE}"
if [[ -n "${VEIL_PROFILE_PATH:-}" && -f "${VEIL_PROFILE_PATH}" ]]; then
  # shellcheck disable=SC1090
  source "${VEIL_PROFILE_PATH}"
fi
set +a

PIPELINE_SCALE="${PIPELINE_WORKER_SCALE:-1}"
INGEST_SCALE="${INGEST_WORKER_SCALE:-1}"
BUILD_FLAG=()
[[ "${COMPOSE_BUILD}" == "1" ]] && BUILD_FLAG=(--build)

case "${ACTION}" in
  up)
    compose up -d "${BUILD_FLAG[@]}" crawl-db nats neo4j graph-bootstrap
    compose up -d "${BUILD_FLAG[@]}" --scale "pipeline_worker=${PIPELINE_SCALE}" \
      --scale "ingest_worker=${INGEST_SCALE}" \
      scrape_worker pipeline_worker ingest_worker api
    if [[ "${ENABLE_ENGAGE}" == "1" ]]; then
      compose up -d "${BUILD_FLAG[@]}" engage-api
      if [[ "${ENABLE_ENGAGE_EVENTS}" == "1" ]]; then
        compose up -d "${BUILD_FLAG[@]}" engage-events-worker
      fi
    fi
    ;;
  down)
    compose down -v --remove-orphans || true
    ;;
  *)
    echo "usage: $0 up|down" >&2
    exit 1
    ;;
esac
