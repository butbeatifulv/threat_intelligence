# Guidance for automated agents (Cursor, CI bots, etc.)

**Core rules (DRY hub):** `make rules-link` from cxado meta-repo → [.cursor/rules/core-karpathy-guidelines.mdc](.cursor/rules/core-karpathy-guidelines.mdc), [core-agent-critic.mdc](.cursor/rules/core-agent-critic.mdc), [core-parallel-branches.mdc](.cursor/rules/core-parallel-branches.mdc), [core-kaizen.mdc](.cursor/rules/core-kaizen.mdc), [core-agent-documentation.mdc](.cursor/rules/core-agent-documentation.mdc). Generic skill: [cxado-skills/agent/karpathy-guidelines](https://github.com/butbeautifulv/cxado-skills). Upstream reference: [refs/andrej-karpathy-skills-main/](../../../../refs/andrej-karpathy-skills-main/) at cxado meta root (do not edit).

**Veil overlay:** [veil-agent-workflow.mdc](.cursor/rules/veil-agent-workflow.mdc) — Go layers (discovery, pipeline, knowledge), `make test-*`.

**Pentest execution** lives in **[veneno](https://github.com/butbeautifulv/veneno)** — not in this repo. See [veil-veneno-split.md](docs/architecture/veil-veneno-split.md).

**Rules index (core + overlay):** [docs/agents/cursor-rules-index.md](docs/agents/cursor-rules-index.md) — catalog, orchestrator vs implementer, adaptation matrix.

## Agent chain (summary)

| Step | Rule / doc |
|------|------------|
| Plan | Master + phase plan in `.cursor/plans/` |
| Implement | [core-parallel-branches.mdc](.cursor/rules/core-parallel-branches.mdc), [veil-agent-workflow.mdc](.cursor/rules/veil-agent-workflow.mdc) |
| Review | [core-agent-critic.mdc](.cursor/rules/core-agent-critic.mdc) |
| Subagents | [veil-agent-subagents.mdc](.cursor/rules/veil-agent-subagents.mdc), [`.cursor/agents/manifest.yaml`](.cursor/agents/manifest.yaml) |
| Merge | Prompt merge to `main` ([core-parallel-branches.mdc](.cursor/rules/core-parallel-branches.mdc) § Merge discipline) |
| Document | [core-agent-documentation.mdc](.cursor/rules/core-agent-documentation.mdc) — includes **README.md**, **CONTRIBUTING.md**, **`.github/repo-description.txt`** |
| Security frameworks | [veil-agent-security-frameworks.mdc](.cursor/rules/veil-agent-security-frameworks.mdc), [docs/external/external-security-frameworks.md](docs/external/external-security-frameworks.md) |
| Cyber playbooks (read) | [docs/playbooks/external-cybersecurity-skills.md](docs/playbooks/external-cybersecurity-skills.md), [docs/architecture/cyber-domain-model.md](docs/architecture/cyber-domain-model.md) — `make corpus-import`, `make skills-index`, veil-mcp `playbook_*` |
| Veneno (pentest) | [veneno](https://github.com/butbeautifulv/veneno) — tool catalog, MCP exec; veil ingests `engage.events` only |
| Archived plans | [.cursor/plans/archive/README.md](.cursor/plans/archive/README.md) — platform P6–P12 |
| Finish | This file § End-of-task checklist |

## Before you change code

1. **Read and follow [docs/agents/coding-style.md](docs/agents/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; three isolated contexts (`discovery/`, `pipeline/`, `knowledge/`); shared wire types in `pkg/`. Pentest execution: [veneno](https://github.com/butbeautifulv/veneno).
2. **Do not add root `go.work`** or cross-layer Go imports between `discovery/`, `pipeline/`, `knowledge/`. Integrate via NATS; all layers may import `pkg/*`.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/architecture/threatintel-runtime.md](docs/architecture/threatintel-runtime.md)**, **[docs/contracts/ingest-contract.md](docs/contracts/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**.
5. Versions: **[versions.env](versions.env)** is the single source of truth for `APP_VERSION` and `GRAPH_PACK_VERSION`.

Reference modules: [discovery/harvest/internal/sources/ti/](discovery/harvest/internal/sources/ti/), [knowledge/ingest/internal/sources/ti/](knowledge/ingest/internal/sources/ti/), [pipeline/ned/internal/sources/ti/](pipeline/ned/internal/sources/ti/).

## Planning and commit rhythm (required for multi-phase work)

Keep diffs reviewable: **one git commit per completed phase or slice**, not one giant commit at the end.

1. **Master plan** — before coding, write or update a master plan (status table with **phase / branch / status / owner**, dependencies). Active plans live in `.cursor/plans/`; completed phases go to `.cursor/plans/archive/` (see [archive/README.md](.cursor/plans/archive/README.md)).
2. **Phase plan** — for the active phase only, add or open a slice plan derived from the master plan (scope, files, acceptance).
3. **Branch per stream** — implementers work on `engage/phase-<NN>-<slug>` (or `feat/<layer>-phase-<NN>-<slug>`), not directly on `main` when multiple agents run in parallel. See [.cursor/rules/core-parallel-branches.mdc](.cursor/rules/core-parallel-branches.mdc).
4. **Execute one phase** — implement only what that phase plan covers; run tests for touched layers.
5. **Commit on the branch** — `git add` + commit like `feat(engage): Phase N — <short title>`; `git push -u origin HEAD`; open a PR to `main`.
6. **Critic gate** — the **orchestrator / main agent session** acts as critic & compliance ([.cursor/rules/core-agent-critic.mdc](.cursor/rules/core-agent-critic.mdc)): plan scope, architecture, tests, graph version; verdict APPROVE / REQUEST_CHANGES before merge.
7. **Merge to `main` promptly** — after critic APPROVE, merge and `git push origin main` so the repo does not drift across parallel branches. See [core-parallel-branches.mdc](.cursor/rules/core-parallel-branches.mdc) § Merge discipline.
8. **Update master plan** — on merge, mark phase `done`, note merge commit SHA; clear or archive branch name.
9. **Actualize documentation** — plans, **[README.md](README.md)**, **[CONTRIBUTING.md](CONTRIBUTING.md)**, **[.github/repo-description.txt](.github/repo-description.txt)** (`make sync-github-metadata`), runtime/deploy docs, parity matrices per [core-agent-documentation.mdc](.cursor/rules/core-agent-documentation.mdc); list touched doc paths in the commit or PR.

If the user asks to “stage all” or catch up after many phases, still document phase boundaries in the commit message body.

### Parallel agents (summary)

| Role | Branch | Merge to `main` |
|------|--------|-----------------|
| Implementer (Task / subagent / second chat) | `engage/phase-NN-slug`, `platform/p0-*` | Only after critic APPROVE; do not start next phase until prior merge is on `main` |
| Critic & compliance (default for orchestrator chat) | `main` | Merges approved branches, pushes `main`, then starts next phase |

Independent phases may run on **different branches at the same time** only if merges keep pace; otherwise **serialize merges** to avoid divergence. Serial phases rebase onto `main` after dependencies merge.

## End-of-task checklist (required)

Complete every step that applies before you consider the task done:

1. **Tests** — run layer targets: `make test-discovery`, `make test-pipeline`, `make test-knowledge` for touched layers. Veneno ingest bridge: `make test-engage-events-pipeline`. Graph read smoke: `make test-graph-read-smoke`.
2. **Graph version** — if you changed ingest-affecting paths (`discovery/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `knowledge/ingest/internal/sources/` including `engage/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`), run `./scripts/release/bump-graph-version.sh patch` and rebuild/publish the graph pack when a new ZIP is needed.
3. **Pre-commit check** — `./scripts/release/check-graph-version-bump.sh` (or `make check-graph-version`).
4. **Commit** — descriptive message (what changed and why). Do not commit secrets or `data/`. Use `git add -A -- . ':!data'` when `data/` causes permission errors. Exclude `**/__pycache__/`.
5. **Push** — `git push origin HEAD` unless the user explicitly forbade push or there is no remote.
6. **GitHub description** — if [.github/repo-description.txt](.github/repo-description.txt) changed, run `make sync-github-metadata`.

## Graph pack releases

- Default version: see [versions.env](versions.env).
- Workflow: [docs/contracts/graph-pack.md](docs/contracts/graph-pack.md).
- Publish: `GRAPH_PACK_VERSION=vX.Y.Z ./scripts/release/publish-graph-pack.sh`.
