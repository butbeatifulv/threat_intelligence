# Cleanup R0 inventory (2026-05-20)

Generated during [veil_cleanup_domain_pkg_master.plan.md](../.cursor/plans/veil_cleanup_domain_pkg_master.plan.md). Regenerate script list: `./scripts/housekeeping/audit-repo-refs.sh` (heuristic; **Makefile** is source of truth).

## Go — removed (2026-05 cleanup)

| Path | Notes |
|------|-------|
| ~~`engage/serve/cmd/browser-agent/`~~ | Removed; browser in `discovery/cmd/browser-agent` |
| ~~`pipeline/pkg/ti/normalize/`~~ | Removed; NED uses `pkg/ti/normalize` |

## Go — keep (legacy profile)

| Path | Notes |
|------|-------|
| `pkg/exec/sandbox.go` | Used when `ENGAGE_EXECUTION_PROFILE` ≠ `client-native` and docker sandbox enabled |

## Scripts — removed

| Path | Notes |
|------|-------|
| ~~`scripts/test/smoke-scrape-e2e.sh`~~ | Removed; use `smoke-discovery-e2e.sh` |

## Scripts — KEEP (Makefile / CI / docs)

All other `scripts/**/*.sh` and `scripts/**/*.py` referenced from [Makefile](../Makefile), [.github/workflows/](../.github/workflows/), or [scripts/README.md](../scripts/README.md). Notable doc-only ops scripts (no Makefile): `compose-scale-veil.sh`, `compose-up-veil-engage.sh`, `migrate-var-veil.sh`, `verify-nvd-enrichment.sh`, `smoke-neo4j-cluster.sh`, `profile-full-enrich.sh`.

## Docs — slim (B4)

| Action | Files |
|--------|-------|
| Keep canonical | `threatintel-runtime.md`, `ingest-contract.md`, `engage-tools.md`, `platform-unified-access.md` |
| Already hubbed | `engage-lab-pentest.md` |
| Pointer only | `engage-client-native-viability.md` → `engage-client-dependencies.md` |

## eval/results policy (B1)

- Track in git: `veil-pentest-*-latest.md`, `veil-pentest-prod-latest.md`, stamps
- Ignore: `hexstrike-*.json`, dated `veil-pentest-*.json` (see `eval/results/.gitignore`)

## Makefile aliases to remove (B2)

- `test-scrape`, `test-scrape-p7c` → use `test-discovery*`
