# openJiuwen Agent Store (reference only)

[Agent Store](https://gitcode.com/openJiuwen/agent-store) is a **catalog of community agent implementations** (Python workflows, finance bots, vibe-coding helpers). In the cxado monorepo, patterns live under `refs/agent-store/` at meta root (`../../../../refs/agent-store/` from this doc).

```bash
make external-clone-agent-store
# or: ./scripts/external/clone-agent-store.sh
```

`.external/` is **gitignored** — clones are local; CI does not require agent-store.

## What Agent Store is

| Area | Content | Veil stance |
|------|---------|-------------|
| `community/` | Standalone agents (deps, API keys, arbitrary code) | **Do not run** in Veil CI/prod stacks |
| `templates/` | `metadata.json` schema for coded/no-code agents | **Inform** Veil manifest fields only |
| `archived/` | Unmaintained samples | Reference only |

## Critical adoption map

| Agent Store idea | Adopt in Veil? | Veil equivalent |
|------------------|----------------|-----------------|
| Per-agent `metadata.json` (id, tags, category) | **Partially** | [`.cursor/agents/manifest.yaml`](../.cursor/agents/manifest.yaml) — orchestrator subagents, not third-party runtimes |
| Community agents as deployable units | **No** | Veil agents are **MCP clients** over `veil-graph` + `veil-engage`, with RBAC and hardening |
| Multi-agent orchestration inside one repo | **No copy-paste** | Veil uses Cursor Task/subagent manifest + critic gate ([AGENTS.md](../AGENTS.md)) |
| Embedded benchmarks (e.g. mango MMLU/GSM8K) | **No** | Use **[GAIA](agent-evaluation-gaia.md)** for general-assistant eval; engage parity for security tools |
| Arbitrary `pip install` + `.env` secrets per agent | **No** | Secrets via deploy profiles; no per-agent dependency sprawl in monorepo |

## Risks if treated naively

1. **Supply chain** — community folders pull diverse PyPI deps; unsuitable for secured engage images.
2. **Scope creep** — agents assume full shell/browser; conflicts with `ENGAGE_DENY_RAW_COMMAND` and catalog-only tools.
3. **No Veil layering** — agents do not respect scrape/pipeline/graph/engage boundaries or graph-version discipline.
4. **Evaluation mismatch** — store demos optimize for demos, not reproducible GAIA-style scoring.

## What we actually extract

- **Catalog metadata shape** (`schema_version`, `id`, `tags`, `category`) when extending [manifest.yaml](../.cursor/agents/manifest.yaml).
- **Documentation pattern** — one README + metadata per capability; Veil mirrors this in `docs/` + phase plans, not in `community/`.
- **Counter-example** — why Veil centralizes tool execution in **engage** instead of per-agent subprocess trees.

## Related

- [agent-evaluation-gaia.md](agent-evaluation-gaia.md) — GAIA benchmark for agent quality metrics
- [mcp-agents.md](../agents/mcp-agents.md) — supported agent integration path
- [external-security-frameworks.md](../external/external-security-frameworks.md) — JCSF/DAF/OWASP (security, not agent catalog)
