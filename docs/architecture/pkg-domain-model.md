# pkg/domain meta-layer (Veil)

Veil domain types today live in per-source packages under `pkg/*/domain`. This document defines the **three contours** that partition those packages, the planned **`pkg/domain` meta-layer** (wave 1 = taxonomy and shared primitives only — no entity migration), and how playbook corpus artifacts split between committed mappings and the upstream skills mirror.

**Related SOT:**

- Package map and layer adapters: [domain-contour.md](../architecture/domain-contour.md)
- Playbook subdomain → graph category taxonomy: [cyber-domain-model.md](../architecture/cyber-domain-model.md)

## Three contours

Veil separates shared domain logic by **runtime role**, not by MITRE vs TI vs vuln subject matter.

| Contour | Flow | Purpose |
|---------|------|---------|
| **Ingest** | scrape → pipeline NED → graph ingest | Normalize external feeds (STIX, advisories, templates, rules) into Neo4j via `harvest` / `commit` envelopes |
| **Engage** | active testing / scan workflows | Target subject, tool catalog specs, findings, jobs — no Neo4j entity structs in engage layer |
| **Knowledge / Playbook** | read-only via veil-api / MCP | Procedure text, framework alignment, skill index — **not** STIX ingest |

```text
Ingest:     discovery/harvest → pipeline/ned → knowledge/ingest  (pkg/*/domain + harvest/commit)
Engage:     engage/serve                                      (pkg/engage/domain/*)
Knowledge:  knowledge/serve                                   (pkg/playbook/*; graph read DTOs stay in knowledge/)
```

Existing per-source packages remain the **entity SOT** until later migration phases. `pkg/domain` adds cross-contour vocabulary only.

## Contour → pkg paths

| Contour | Package paths | Wire / transport | Notes |
|---------|---------------|------------------|-------|
| **Ingest** | `pkg/ti/domain`, `pkg/vuln/domain`, `pkg/lola/domain`, `pkg/ds/domain`, `pkg/sbom/domain`, `pkg/nuclei/domain`, `pkg/coderules/domain` | `pkg/harvest`, `pkg/commit` | Rich entities (IOC, Vulnerability, …) and thin refs (AdvisoryRef, Template, Resource, RuleFile) |
| **Ingest** (pipeline-only) | `pkg/ti/normalize`, `pkg/ti/validate`, `pkg/ti/ids` | commit payload | Normalization runs in NED only — graph ingest does not re-normalize |
| **Engage** | `pkg/engage/domain/target`, `pkg/engage/domain/tool`, `pkg/engage/domain/report`, `pkg/engage/domain/job` | `pkg/engage/contract`, `pkg/engage/events` | Engage HTTP/MCP DTOs; catalog YAML stays in `engage/serve/catalog/` |
| **Knowledge / Playbook** | `pkg/playbook/domain`, `pkg/playbook/index`, `pkg/playbook/procedure`, `pkg/playbook/framework`, `pkg/playbook/cataloglink` | veil-api / MCP | `cataloglink` is the only cross-`pkg` runtime import from playbook → engage catalog names |
| **Knowledge / Playbook** (corpus) | `pkg/playbook/corpus/mappings` | framework routes | MITRE Navigator layer, NIST CSF, OWASP — Veil SOT (see below) |
| **Decision** (adjacent) | `pkg/decision` | effectiveness tables | Tool selection; optional boost from playbook via veil-api — not a fourth contour |

Layer adapters (scrape, NED, graph ingest, engage serve) stay outside `pkg/`; see [domain-contour.md](../architecture/domain-contour.md) § Layer adapters.

## Playbook corpus split

Playbook material is **two committed artifacts** plus a dev import path:

| Artifact | Path | Role |
|----------|------|------|
| **Framework mappings (Veil SOT)** | [pkg/playbook/corpus/mappings/](../pkg/playbook/corpus/mappings/) | ATT&CK Navigator layer (Enterprise v14 in layer file), NIST CSF 2.0, OWASP alignment — consumed by `pkg/playbook/framework` and veil-api framework routes |
| **Procedure bodies (upstream mirror)** | [corpus/anthropic-cybersecurity-skills/skills/](../corpus/anthropic-cybersecurity-skills/skills/) | agentskills.io-style `SKILL.md` mirror (754 skills); `CorpusPath` in [docs/skills-index/cyber-skills.json](../skills-index/cyber-skills.json) |
| **Native procedures (Veil curated)** | [pkg/playbook/procedures/](../pkg/playbook/procedures/) | Structured YAML; `GetSpec` prefers native over corpus parse |
| **Generated index** | [docs/skills-index/cyber-skills.json](../skills-index/cyber-skills.json) | Metadata + `corpus_path`; not ingest STIX |
| **Dev import** | `.external/Anthropic-Cybersecurity-Skills-main/` | Gitignored; `make corpus-import` only |

`pkg/playbook/corpus/` holds **mappings and attribution only** — not the SKILL.md tree. Procedure narrative lives under repo-root `corpus/anthropic-cybersecurity-skills/`. See [pkg/playbook/corpus/README.md](../pkg/playbook/corpus/README.md) and [cyber-domain-model.md](../architecture/cyber-domain-model.md).

## Planned `pkg/domain` primitives (P1+)

Wave 1 introduces `pkg/domain` as a **meta-layer**: enums, refs, and type aliases — **no I/O**, no changes to harvest/commit JSON wire.

| Primitive | Purpose | Covers today |
|-----------|---------|--------------|
| **`Contour`** | `Ingest` \| `Engage` \| `Knowledge` | Documents and CI boundaries; Neo4j entity types belong to Ingest, not Engage |
| **`Source`** | Stable source id string + registry | Mirrors `harvest.Source*` / `commit.Source*` (`ti`, `vuln`, `lola`, `ds`, `sbom`, `nuclei`, `coderules`, …) |
| **`SourceRef`** | `{Source, Key, Path, Kind}` | Thin identity pattern in `AdvisoryRef`, `Template`, `Resource`, `RuleFile` — adapters in later phases |
| **`VeilCategory`** | Graph/read category + subdomain map | Aligns `pkg/playbook/framework/subdomain.go` with [cyber-domain-model.md](../architecture/cyber-domain-model.md) § Subdomain → Veil categories |

Planned file layout (implementation deferred to P1–P6):

```text
pkg/domain/
  doc.go           # package comment → this doc
  contour.go       # Contour enum
  source.go        # Source + Valid(), AllSources()
  ref.go           # SourceRef + helpers
  taxonomy.go      # VeilCategory, subdomain family map
  ingest.go        # ToSourceRef adapters (thin domain types)
  knowledge.go     # type aliases to pkg/playbook/domain; FrameworkContour metadata
  engage.go        # type aliases to pkg/engage/domain/*
```

**Explicitly out of scope for wave 1:** `harvest.Envelope`, `commit.Envelope`, Neo4j projections (`knowledge/connector/query`), NVD parse, engage catalog YAML, moving `IOC` or other rich structs out of `pkg/ti/domain`.

**Import rule:** `pkg/domain` may import `pkg/*/domain` and `pkg/playbook/domain`; it must not import `harvest`, `commit`, or `knowledge/serve` (avoids cycles).

## Phase map

| Phase | Deliverable |
|-------|-------------|
| **P0** (this doc) | Contour model + corpus split SOT |
| P1 | `pkg/domain` scaffold (`contour`, `source`, `ref`, `taxonomy`) + tests |
| P2 | `SourceRef` adapters for thin ingest types |
| P3 | Source registry vs harvest/commit consts |
| P4 | Knowledge aliases + framework contour metadata |
| P5 | Engage type aliases |
| P6 | Makefile gates, manifest phases, AGENTS.md |

See [.cursor/plans/pkg_domain_umbrella_113018bf.plan.md](../.cursor/plans/pkg_domain_umbrella_113018bf.plan.md) for full wave plan.
