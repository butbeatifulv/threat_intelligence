#!/usr/bin/env bash
# Enforce pkg unit-test tiers: T0 (tests present), T2 (coverage floors).
# Usage: ./scripts/test/pkg-cover.sh
# Requires: go test from repo root; respects env -u GOWORK per module.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if [[ "${PKG_COVER_STRICT:-}" == "1" ]]; then
	DEFAULT_FLOOR=100
	MCP_FLOOR=100
	EXEC_FLOOR=100
	KEYCLOAK_FLOOR=100
	_LOG_PREFIX="[pkg-cover-strict]"
else
	DEFAULT_FLOOR=70
	MCP_FLOOR=60
	EXEC_FLOOR=50
	KEYCLOAK_FLOOR=45
	_LOG_PREFIX="[pkg-cover]"
fi

# T0: data-only structs — presence tests only ([no statements] OK).
types_only=(
	github.com/butbeautifulv/veil/pkg/coderules/domain
	github.com/butbeautifulv/veil/pkg/ds/domain
	github.com/butbeautifulv/veil/pkg/lola/domain
	github.com/butbeautifulv/veil/pkg/nuclei/domain
	github.com/butbeautifulv/veil/pkg/sbom/domain
	github.com/butbeautifulv/veil/pkg/vuln/domain
	github.com/butbeautifulv/veil/pkg/playbook/domain
	github.com/butbeautifulv/veil/pkg/engage/contract
	github.com/butbeautifulv/veil/pkg/engage/domain/job
	github.com/butbeautifulv/veil/pkg/engage/domain/target
)

is_types_only() {
	local pkg="$1"
	for t in "${types_only[@]}"; do
		[[ "${pkg}" == "${t}" ]] && return 0
	done
	return 1
}

floor_for() {
	local pkg="$1"
	case "${pkg}" in
	github.com/butbeautifulv/veil/pkg/mcp) echo "${MCP_FLOOR}" ;;
	github.com/butbeautifulv/veil/pkg/exec) echo "${EXEC_FLOOR}" ;;
	github.com/butbeautifulv/veil/pkg/auth/keycloak) echo "${KEYCLOAK_FLOOR}" ;;
	*) echo "${DEFAULT_FLOOR}" ;;
	esac
}

log() { printf '%s %s\n' "${_LOG_PREFIX}" "$*"; }
fail() { log "FAIL: $*"; exit 1; }

check_notest() {
	local dir="$1"
	local mod_label="$2"
	local missing
	missing="$(cd "${dir}" && env -u GOWORK go list -f '{{if not .TestGoFiles}}{{if .GoFiles}}{.ImportPath}}{{end}}{{end}}' ./... 2>/dev/null | grep -v '^$' || true)"
	if [[ -n "${missing}" ]]; then
		fail "${mod_label}: packages without *_test.go:\n${missing}"
	fi
}

declare -a COVER_LINES=()

run_cover() {
	local dir="$1"
	local label="$2"
	check_notest "${dir}" "${label}"
	while IFS= read -r line; do
		COVER_LINES+=("${line}")
	done < <(cd "${dir}" && env -u GOWORK go test ./... -cover 2>&1 | grep -E '^(ok|FAIL)\s+' || true)
}

run_cover "${ROOT}/pkg" "pkg"
run_cover "${ROOT}/pkg/engage" "pkg/engage"
run_cover "${ROOT}/pkg/api" "pkg/api"
run_cover "${ROOT}/pkg/auth" "pkg/auth"
run_cover "${ROOT}/pkg/mcp" "pkg/mcp"
run_cover "${ROOT}/pkg/exec" "pkg/exec"

for line in "${COVER_LINES[@]}"; do
	[[ "${line}" == FAIL* ]] && fail "tests failed: ${line}"
	pkg="$(echo "${line}" | awk '{print $2}')"
	rest="$(echo "${line}" | cut -d$'\t' -f3-)"
	if [[ "${rest}" == *"[no statements]"* ]]; then
		if ! is_types_only "${pkg}"; then
			fail "${pkg}: [no statements] but not in types-only allowlist"
		fi
		log "T0 ${pkg} (no statements)"
		continue
	fi
	if [[ "${rest}" != *"coverage:"* ]]; then
		continue
	fi
	pct="$(echo "${rest}" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p')"
	if [[ -z "${pct}" ]]; then
		fail "${pkg}: could not parse coverage from: ${line}"
	fi
	floor="$(floor_for "${pkg}")"
	awk -v p="${pct}" -v f="${floor}" 'BEGIN { exit (p+0 >= f+0) ? 0 : 1 }' || fail "${pkg}: ${pct}% < floor ${floor}%"
	if [[ "${PKG_COVER_STRICT:-}" == "1" ]]; then
		log "T3 ${pkg} ${pct}% (floor ${floor}%)"
	else
		log "T2 ${pkg} ${pct}% (floor ${floor}%)"
	fi
done

if [[ "${PKG_COVER_STRICT:-}" == "1" ]]; then
	log "OK — pkg T3 coverage gates passed"
else
	log "OK — pkg coverage gates passed"
fi
