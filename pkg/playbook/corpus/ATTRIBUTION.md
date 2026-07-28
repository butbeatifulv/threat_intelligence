# Corpus attribution

Vendored from the community **[Anthropic Cybersecurity Skills](https://github.com/anthropics/anthropic-cybersecurity-skills)** repository (not affiliated with Anthropic PBC).

| Field | Value |
|-------|--------|
| Upstream SHA | See [VERSION](VERSION) |
| License | Apache-2.0 (repository [LICENSE](../../../.external/Anthropic-Cybersecurity-Skills-main/LICENSE) at import time); per-skill licenses in each `SKILL.md` frontmatter |
| Veil layout | `mappings/` → this tree; `skills/` → [corpus/anthropic-cybersecurity-skills/skills/](../../../corpus/anthropic-cybersecurity-skills/skills/) |

Veil may evolve **mappings** and domain docs under `pkg/playbook/corpus/`; upstream skill prose stays mirrored until explicitly marked `veil_derived` in frontmatter.

Regenerate from a local clone: `make corpus-import`.
