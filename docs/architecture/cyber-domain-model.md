# Cyber domain model (Veil)

Veil treats the Anthropic Cybersecurity Skills corpus as a **procedure + taxonomy** layer, separate from executable engage tools and from official MITRE STIX ingest.

## Layout (committed)

| Artifact | Path | Notes |
|----------|------|--------|
| Framework mappings | [pkg/playbook/corpus/mappings/](../pkg/playbook/corpus/mappings/) | Navigator layer (ATT&CK v14 in layer file), NIST CSF 2.0, OWASP — **Veil SOT** |
| Agent playbooks | [corpus/anthropic-cybersecurity-skills/skills/](../corpus/anthropic-cybersecurity-skills/skills/) | agentskills.io-style `SKILL.md` mirror |
| Index | [docs/skills-index/cyber-skills.json](../skills-index/cyber-skills.json) | Generated metadata + `corpus_path` |
| Dev import source | `.external/Anthropic-Cybersecurity-Skills-main/` | Gitignored; `make corpus-import` only |

## Subdomain → Veil categories

Anthropic skills use **26 subdomains** (e.g. `digital-forensics`, `threat-hunting`). Veil maps them to existing graph/read categories:

| Subdomain family | Veil category | Relationship |
|------------------|---------------|--------------|
| digital-forensics, incident-response | `playbook` + `mitre` | Procedure text; techniques from STIX ingest |
| threat-hunting, detection-engineering | `playbook` + `detection` | Procedures; Sigma/YARA in `detection` |
| web-application-security, penetration-testing | `playbook` + engage catalog | Procedures recommend catalog tools, not new MCP tools |
| threat-intelligence | `playbook` + `ti` | OSINT/MISP procedures ≠ IOC graph |
| compliance-governance, soc-operations | `playbook` + [external-security-frameworks.md](../external/external-security-frameworks.md) | Aligns with NIST CSF tables in mappings |

## MITRE ATT&CK version note

- **Navigator layer** in mappings targets **ATT&CK Enterprise v14** (see [mappings/README.md](../pkg/playbook/corpus/mappings/README.md)).
- **Neo4j** `AttackTechnique` nodes come from **enterprise STIX** (LOLA harvest; matrix version may differ, e.g. v18).
- Veil joins on **technique id string** (`T1059.001`), not matrix version objects.

## Framework files

| Framework | File(s) |
|-----------|---------|
| MITRE ATT&CK | [attack-navigator-layer.json](../pkg/playbook/corpus/mappings/attack-navigator-layer.json), [mitre-attack/coverage-summary.md](../pkg/playbook/corpus/mappings/mitre-attack/coverage-summary.md) |
| NIST CSF 2.0 | [nist-csf/README.md](../pkg/playbook/corpus/mappings/nist-csf/README.md), [csf-alignment.md](../pkg/playbook/corpus/mappings/nist-csf/csf-alignment.md) |
| OWASP | [owasp/README.md](../pkg/playbook/corpus/mappings/owasp/README.md) |

## Operations

```bash
make corpus-import      # rsync from .external (dev)
make skills-index       # regenerate cyber-skills.json
make check-corpus-mappings
```

See [external-cybersecurity-skills.md](../playbooks/external-cybersecurity-skills.md) for API/MCP tools.

Cross-contour taxonomy (Ingest / Engage / Knowledge): [pkg-domain-model.md](../architecture/pkg-domain-model.md) and [pkg/domain/](../pkg/domain/).

## Decision vs playbook (DRY)

| Question | Layer |
|----------|--------|
| Which **engage catalog tool** to run? | [pkg/decision](../pkg/decision) — effectiveness tables; optional **+0.12 boost** from `playbook_catalog_tools` via veil-api |
| What **procedure** should an agent follow? | [pkg/playbook/procedure](../pkg/playbook/procedure) — structured steps from SKILL.md |
| What **MITRE / CSF** frame applies? | [pkg/playbook/framework](../pkg/playbook/framework) + [mappings](../pkg/playbook/corpus/mappings/) |

Import progress: [playbook-import-matrix.md](playbook-import-matrix.md). Regenerate structured index: `make procedures-index`.
