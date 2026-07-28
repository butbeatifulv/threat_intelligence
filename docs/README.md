# Veil documentation

Markdown docs are grouped by topic. Generated indexes and wire schemas stay at the repo root of `docs/` because tooling and CI depend on fixed paths.

## Categories

| Folder | Scope | Start here |
|--------|--------|------------|
| [architecture/](architecture/) | Platform layers, domain contours, runtime compose, closed-loop pilots | [platform-architecture.md](architecture/platform-architecture.md), [threatintel-runtime.md](architecture/threatintel-runtime.md) |
| [contracts/](contracts/) | NATS ingest envelopes, graph pack workflow | [ingest-contract.md](contracts/ingest-contract.md), [graph-pack.md](contracts/graph-pack.md) |
| [playbooks/](playbooks/) | Anthropic cybersecurity skills corpus (754), API/MCP read path | [external-cybersecurity-skills.md](playbooks/external-cybersecurity-skills.md) |
| [engage/](engage/) | Tool catalog, MCP topology, install, lab pentest, audits | [engage-tools.md](engage/engage-tools.md), [engage-runtime.md](engage/engage-runtime.md) |
| [deploy/](deploy/) | Secure/hybrid deploy, Keycloak, Neo4j cluster | [deploy-secure.md](deploy/deploy-secure.md) |
| [agents/](agents/) | MCP setup, coding style, GAIA evaluation | [mcp-agents.md](agents/mcp-agents.md), [coding-style.md](agents/coding-style.md) |
| [external/](external/) | Legacy HexStrike reference, security framework index | [external-security-frameworks.md](external/external-security-frameworks.md) |
| [development/](development/) | `pkg/` test coverage matrix, cleanup inventory | [pkg-test-coverage.md](development/pkg-test-coverage.md) |

## Root artifacts (do not move)

| Path | Role |
|------|------|
| [skills-index/](skills-index/) | Generated `cyber-skills.json`, `procedures-index.json` |
| [schemas/](schemas/) | `harvest` / `commit` JSON schemas |
| [templates/](templates/) | Graph pack release notes template |
| [openapi.yaml](openapi.yaml) | veil-api OpenAPI |
| [graph-pack-manifest.schema.json](graph-pack-manifest.schema.json) | Pack manifest validation |
| [assets/veil.png](assets/veil.png) | README banner |

## Quick links

- Project overview: [README.md](../README.md)
- Agents / PR workflow: [AGENTS.md](../AGENTS.md)
- Deploy compose: [deploy/README.md](../deploy/README.md)
- Engage layer: [engage/README.md](../engage/README.md)
