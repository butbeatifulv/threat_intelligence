# External security frameworks (reference layer)

Veil **does not ship** full JCSF/DAF spreadsheets in `refs/` — those directories are **read-only reference** (symlink to [cxado-references](https://github.com/butbeautifulv/cxado-references) via meta-repo). Operational truth lives in:

| Artifact | Role |
|----------|------|
| [deploy/security/veil-controls.yaml](../deploy/security/veil-controls.yaml) | Machine-readable control catalog + framework IDs |
| [docs/engage/engage-hardening.md](engage-hardening.md) | Engage active-defense hardening |
| [docs/engage/engage-agentic-threats.md](engage-agentic-threats.md) | Agentic AI / MCP threat mapping (OWASP + DAF MLSO) |
| [docs/deploy/deploy-platform-hybrid.md](../deploy/deploy-platform-hybrid.md) | Container platform (JCSF-aligned) |

## Vendored references

| Path | Framework | What we extract |
|------|-----------|-----------------|
| [refs/Jet-Container-Security-Framework-main/](../refs/Jet-Container-Security-Framework-main/) | **JCSF** (Jet) | Five domains: nodes, runtime, images, manifests, container runtime → P5 Terraform/Ansible/Helm |
| [refs/DevSecOps-Assessment-Framework-main/](../refs/DevSecOps-Assessment-Framework-main/) | **DAF** | Secure SDLC practices (secrets, CI/CD, container images, SCM) → CI gates, profiles |
| `DAF_MLSO_public_RU.md` in same tree | **DAF MLSO** | AI/agent/MCP least privilege → engage runner + MCP |
| [refs/owasp/Cheat-Sheet-Agentic-AI-Solution-Landscape-Q226-1-1.pdf](../refs/owasp/Cheat-Sheet-Agentic-AI-Solution-Landscape-Q226-1-1.pdf) | **OWASP GenAI** | Agentic lifecycle SecOps → MCP/tooling controls |
| [refs/owasp/Cheat-Sheet-Red-Teaming-AI-Solution-Landscape-Q226.pdf](../refs/owasp/Cheat-Sheet-Red-Teaming-AI-Solution-Landscape-Q226.pdf) | **OWASP Red Team AI** | Test/evaluate phase → self-test scope (no host attacks) |
| [Карта инструментов DevSecOps.pdf](../refs/Карта%20инструментов%20DevSecOps.pdf) | Tool landscape | Informative; Veil **implements** tools via engage catalog, not every box on the map |
| [refs/agent-store/](../refs/agent-store/) | **openJiuwen Agent Store** | Catalog/metadata patterns only — see [external-agent-store.md](external-agent-store.md) |
| [.external/Anthropic-Cybersecurity-Skills-main/](../.external/Anthropic-Cybersecurity-Skills-main/) | **Cybersecurity Skills** (community) | Procedure playbooks for agents — see [external-cybersecurity-skills.md](../playbooks/external-cybersecurity-skills.md); index in [skills-index/](../skills-index/) |
| [GAIA arXiv:2311.12983](https://arxiv.org/abs/2311.12983) | **GAIA benchmark** | General-assistant agent eval (HF optional) — [agent-evaluation-gaia.md](agent-evaluation-gaia.md) |

## Adoption principles (critical)

1. **Do not copy controls blindly** — JCSF L4 practices may be incompatible with lab/smoke profiles (`smoke-minimal`, local runner).
2. **Separate roles** — graph read (`veil-mcp`) vs tool exec (`veil-engage`) vs batch scrape; each maps to different DAF/JCSF domains.
3. **Secured engage infra** — production engage uses `secure-engage.env` + Docker runner + `ENGAGE_TARGET_GUARD=block` for SSRF-style abuse from agents.
4. **`refs/` is never executed** — same rule as HexStrike reference; no CI dependency on PDF binaries. `.external/Anthropic-Cybersecurity-Skills-main/` remains Veil-local for `make corpus-import` only.

## Automated verification

```bash
make test-engage-hardening          # unit + compose + veil-controls audit
python3 scripts/engage/hardening-framework-audit.py
make test-agent-eval-pilot          # GAIA offline pilot (no HF token)
```

## Maturity targets (Veil-specific)

| Area | Target level | Rationale |
|------|--------------|-----------|
| Engage prod overlay | JCSF **L2** / DAF Kirill **2–3** | Runner isolation, auth, audit, CI gates |
| Graph secure read | JCSF manifests **L2** | TLS edge, auth required |
| Full stack scrape/pipeline | DAF **L1–2** | SBOM/graph pack checksum; image scan = roadmap |
| Agentic MCP | DAF MLSO **L1–2** | Catalog tools, target guard, no raw shell |

Gaps and roadmap items are tracked in [veil_deploy_platform_p5_hybrid.plan.md](../.cursor/plans/veil_deploy_platform_p5_hybrid.plan.md).
