# Guidance for automated agents (Cursor, CI bots, etc.)

**Behavioral guidelines (Karpathy + Veil):** [.cursor/rules/veil-karpathy-guidelines.mdc](.cursor/rules/veil-karpathy-guidelines.mdc), skill [.cursor/skills/veil-karpathy-guidelines/SKILL.md](.cursor/skills/veil-karpathy-guidelines/SKILL.md). Upstream reference: [refs/andrej-karpathy-skills-main/](refs/andrej-karpathy-skills-main/) (do not edit).

**Metacognition on errors (5 Whys, Gemba Kaizen, 1% improvement):** [.cursor/rules/veil-agent-kaizen-metacognition.mdc](.cursor/rules/veil-agent-kaizen-metacognition.mdc) — mandatory when tests, CI, smokes, or builds fail; document root cause before the next fix.

**Documentation in the agent chain:** [.cursor/rules/veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc) — after each merge, actualize plans, runtime docs, and descriptions the next agent reads; use structured reasoning in phase plans (constraints, few-shot examples from prior phases, explicit `make` DoD).

## Agent chain (summary)

| Step | Rule / doc |
|------|------------|
| Plan | Master + phase plan in `.cursor/plans/` |
| Implement | [veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) |
| Review | [veil-agent-critic.mdc](.cursor/rules/veil-agent-critic.mdc) |
| Subagents | [veil-agent-subagents.mdc](.cursor/rules/veil-agent-subagents.mdc), [`.cursor/agents/manifest.yaml`](.cursor/agents/manifest.yaml) |
| Merge | Prompt merge to `main` ([veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) § Merge discipline) |
| Document | [veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc) — includes **README.md**, **CONTRIBUTING.md**, **`.github/repo-description.txt`** |
| Security frameworks | [veil-agent-security-frameworks.mdc](.cursor/rules/veil-agent-security-frameworks.mdc), [docs/external/external-security-frameworks.md](docs/external/external-security-frameworks.md) |
| Cyber playbooks (read) | [docs/playbooks/external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md), [docs/architecture/cyber-domain-model.md](docs/architecture/cyber-domain-model.md) — `make corpus-import`, `make skills-index`, veil-mcp `playbook_*` |
| Agent evaluation | [docs/agents/agent-evaluation-gaia.md](docs/agents/agent-evaluation-gaia.md) — [arXiv:2311.12983](https://arxiv.org/abs/2311.12983); `make test-agent-eval-pilot` / `test-agent-eval-paper`; HF optional |
| Engage tool coverage | [engage_tools_full_coverage.plan.md](.cursor/plans/engage_tools_full_coverage.plan.md) — legacy parity (**P9f**, 158/158) |
| Repo cleanup (domain/pkg, bloat) | [veil_cleanup_domain_pkg_master.plan.md](.cursor/plans/veil_cleanup_domain_pkg_master.plan.md) — inventory [docs/development/cleanup-r0-inventory.md](docs/development/cleanup-r0-inventory.md) |
| Engage client-native v3 | [engage_mcp_client_native_execution_master.plan.md](.cursor/plans/engage_mcp_client_native_execution_master.plan.md) — **active** |
| Lab pentest & install | [engage-lab-pentest.md](docs/engage/engage-lab-pentest.md), [engage-red-blue-bugs.md](docs/engage/engage-red-blue-bugs.md) |
| Archived plans | [.cursor/plans/archive/README.md](.cursor/plans/archive/README.md) — platform P6–P12, Engage phases, lab tracks |
| Finish | This file § End-of-task checklist |

**Completed program tracks:** indexed under [archive/README.md](.cursor/plans/archive/README.md). Sign-off KPIs: [engage-audit-report.md](docs/engage/engage-audit-report.md). Operator contract (P12): [platform-unified-access.md](docs/architecture/platform-unified-access.md). **Critical fix on main:** catalog merge order `tools.yaml` → `tools.live.yaml` (`634e067`).

**Do not claim “HexStrike execution done” or “full port complete”** until `make test-engage-executable-matrix` is green (**158/158**, branch **P9f**) and docs KPIs match [engage-tools.md](docs/engage/engage-tools.md). Phase 30 sign-off alone is insufficient.

## Before you change code

1. **Read and follow [docs/agents/coding-style.md](docs/agents/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; four isolated contexts (`discovery/`, `pipeline/`, `knowledge/`, `engage/`); domain packages per source; shared wire types in `pkg/`. Before merge, check the [PR checklist](docs/agents/coding-style.md#pr-checklist).
2. **Do not add root `go.work`** or cross-layer Go imports between `discovery/`, `pipeline/`, `knowledge/`, `engage/`. Discovery/pipeline/knowledge integrate via NATS; engage calls knowledge only via HTTP veil-api; all layers may import `pkg/*`. Cross-contour taxonomy and source registry: **[pkg/domain/](pkg/domain/)** — see [docs/architecture/pkg-domain-model.md](docs/architecture/pkg-domain-model.md).
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/architecture/threatintel-runtime.md](docs/architecture/threatintel-runtime.md)**, **[docs/contracts/ingest-contract.md](docs/contracts/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**. Unified edge (P12): **[docs/architecture/platform-unified-access.md](docs/architecture/platform-unified-access.md)** — one TLS host, `/v1/*` + `/api/*` + `/mcp/graph` + `/mcp/engage`; stateless scale `VEIL_*_SCALE` ∈ {4,8,16}; prod Neo4j Enterprise 3-core.
5. Versions: **[versions.env](versions.env)** is the single source of truth for `APP_VERSION` and `GRAPH_PACK_VERSION`.

Reference modules: [discovery/harvest/internal/sources/ti/](discovery/harvest/internal/sources/ti/), [knowledge/ingest/internal/sources/ti/](knowledge/ingest/internal/sources/ti/), [pipeline/ned/internal/sources/ti/](pipeline/ned/internal/sources/ti/).

## Planning and commit rhythm (required for multi-phase work)

Keep diffs reviewable: **one git commit per completed phase or slice**, not one giant commit at the end.

1. **Master plan** — before coding, write or update a master plan (status table with **phase / branch / status / owner**, dependencies). Active plans live in `.cursor/plans/`; completed phases go to `.cursor/plans/archive/` (see [archive/README.md](.cursor/plans/archive/README.md)).
2. **Phase plan** — for the active phase only, add or open a slice plan derived from the master plan (scope, files, acceptance).
3. **Branch per stream** — implementers work on `engage/phase-<NN>-<slug>` (or `feat/<layer>-phase-<NN>-<slug>`), not directly on `main` when multiple agents run in parallel. See [.cursor/rules/veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc).
4. **Execute one phase** — implement only what that phase plan covers; run tests for touched layers.
5. **Commit on the branch** — `git add` + commit like `feat(engage): Phase N — <short title>`; `git push -u origin HEAD`; open a PR to `main`.
6. **Critic gate** — the **orchestrator / main agent session** acts as critic & compliance ([.cursor/rules/veil-agent-critic.mdc](.cursor/rules/veil-agent-critic.mdc)): plan scope, architecture, tests, graph version; verdict APPROVE / REQUEST_CHANGES before merge.
7. **Merge to `main` promptly** — after critic APPROVE, merge and `git push origin main` so the repo does not drift across parallel branches. See [veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) § Merge discipline.
8. **Update master plan** — on merge, mark phase `done`, note merge commit SHA; clear or archive branch name.
9. **Actualize documentation** — plans, **[README.md](README.md)**, **[CONTRIBUTING.md](CONTRIBUTING.md)**, **[.github/repo-description.txt](.github/repo-description.txt)** (`make sync-github-metadata`), runtime/deploy docs, parity matrices per [veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc); list touched doc paths in the commit or PR.

If the user asks to “stage all” or catch up after many phases, still document phase boundaries in the commit message body.

### Parallel agents (summary)

| Role | Branch | Merge to `main` |
|------|--------|-----------------|
| Implementer (Task / subagent / second chat) | `engage/phase-NN-slug`, `platform/p0-*` | Only after critic APPROVE; do not start next phase until prior merge is on `main` |
| Critic & compliance (default for orchestrator chat) | `main` | Merges approved branches, pushes `main`, then starts next phase |

Independent phases may run on **different branches at the same time** only if merges keep pace; otherwise **serialize merges** to avoid divergence. Serial phases rebase onto `main` after dependencies merge.

## End-of-task checklist (required)

Complete every step that applies before you consider the task done:

1. **Tests** — run layer targets from repo root: `make test-discovery`, `make test-pipeline`, `make test-knowledge`, `make test-engage` for the layers you touched. For `knowledge/serve` only: `make test-knowledge-serve` (`-race`). Knowledge read Docker smoke: `make test-graph-read-smoke`. Engage: `make test-engage-parity` when changing catalog; before claiming full execution port, run `make test-engage-executable-matrix` (**P9f**, 158/158). Engage events bus (`engage/.../events`, `pipeline/engage-events`, `knowledge/ingest/.../engage`): also `make test-pipeline`; Docker `make test-engage-events-pipeline`, `make test-engage-veil-stack-ci`. Platform: `make test-platform-p7` (pkg domain + bus, no Docker), `make test-platform-p0` (bus unit tests), `make test-platform-unified-edge` (P12 Docker nginx edge), `make test-platform-closed-loop` (Docker pilot), optional `make test-platform-full-loop` (discovery + engage, heavy) — [docs/architecture/platform-unified-access.md](docs/architecture/platform-unified-access.md) — [docs/architecture/platform-closed-loop-pilot.md](docs/architecture/platform-closed-loop-pilot.md), [docs/architecture/platform-full-loop-smoke.md](docs/architecture/platform-full-loop-smoke.md). Engage hardening (secured infra): `make test-engage-hardening` — [docs/engage/engage-hardening.md](docs/engage/engage-hardening.md).
2. **Graph version** — if you changed ingest-affecting paths (`discovery/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `knowledge/ingest/internal/sources/` including `engage/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`), run `./scripts/release/bump-graph-version.sh patch` and rebuild/publish the graph pack when a new ZIP is needed.
3. **Pre-commit check** — `./scripts/release/check-graph-version-bump.sh` (or `make check-graph-version`).
4. **Commit** — descriptive message (what changed and why). Do not commit secrets or `data/`. Use `git add -A -- . ':!data'` when `data/` causes permission errors. Exclude `**/__pycache__/`.
5. **Push** — `git push origin HEAD` unless the user explicitly forbade push or there is no remote.
6. **GitHub description** — if [.github/repo-description.txt](.github/repo-description.txt) changed, run `make sync-github-metadata` (or rely on [`.github/workflows/docs.yml`](.github/workflows/docs.yml) on push to `main`).

### Engage 158 install quickstart (agent-first)

When working on host-tool readiness for Engage, use this deterministic chain:

1. `make engage-tool-source-map`
2. `make engage-tool-install-coverage`
3. `./scripts/engage/preflight-client-tools.sh --profile recommended --emit-install-plan --policy full-auto`
4. `make engage-install-fallback` or `make engage-install-kali-fallback` (Debian/Ubuntu opt-in)
5. `make test-engage-red-blue`

Core47 parity quick path:

1. `./scripts/ops/install-engage-host-tools.sh --yes --profile core47 --policy full-auto`
2. `./scripts/engage/preflight-client-tools.sh --profile core47 --json`
3. `make test-engage-red-blue`

## Graph pack releases

- Default version: see [versions.env](versions.env).
- Workflow: [docs/contracts/graph-pack.md](docs/contracts/graph-pack.md).
- Publish: `GRAPH_PACK_VERSION=vX.Y.Z ./scripts/release/publish-graph-pack.sh`.
