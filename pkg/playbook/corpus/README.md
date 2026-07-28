# Playbook corpus (Veil)

Committed cybersecurity domain material for agents and veil-api.

| Path | Role |
|------|------|
| [mappings/](mappings/) | **Veil SOT** — MITRE ATT&CK Navigator layer, coverage summaries, NIST CSF 2.0 and OWASP alignment (markdown + JSON) |
| [ATTRIBUTION.md](ATTRIBUTION.md) | Upstream credit and license pointer |
| [VERSION](VERSION) | Pinned upstream SHA and import metadata |

Procedure bodies (754 `SKILL.md` files) live under [corpus/anthropic-cybersecurity-skills/skills/](../../corpus/anthropic-cybersecurity-skills/skills/). Machine-readable summaries: [docs/skills-index/cyber-skills.json](../../docs/skills-index/cyber-skills.json).

## Veil integration

| Consumer | Usage |
|----------|--------|
| `pkg/playbook/index` | Search index + read `SKILL.md` bodies |
| `pkg/playbook/framework` | Parse `mappings/attack-navigator-layer.json` |
| veil-api / veil-mcp | `playbook_*` and framework read tools |
| Neo4j (optional) | `CyberSkill` + `HAS_PLAYBOOK` from index seed |

Human overview of the security domain model: [docs/architecture/cyber-domain-model.md](../../docs/architecture/cyber-domain-model.md).
