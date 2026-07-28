# Veil domain contour (pkg SOT)

Shared threat-intelligence types and rules live under `pkg/`. Runtime layers (`discovery/`, `pipeline/`, `knowledge/`, `engage/`) keep source-specific adapters only.

## pkg/domain meta-layer

Per-source packages below remain entity SOT. A planned top-level [`pkg/domain/`](../pkg/domain/) meta-layer (contours, source registry, shared refs, Veil category taxonomy) is documented in **[pkg-domain-model.md](../architecture/pkg-domain-model.md)** — three contours (Ingest, Engage, Knowledge/Playbook), corpus split, and P1+ primitive layout. No Go code in P0; existing import paths unchanged until later phases.

## Package map

| Package | Role | Imported by |
|---------|------|-------------|
| `pkg/ti/domain` | IOC, Actor, Campaign, Cluster, Report | scrape TI, pipeline NED, graph ingest TI |
| `pkg/vuln/domain` | Vulnerability, CVSS, CPE, ExploitRef | scrape vuln, pipeline NED, graph ingest vuln |
| `pkg/lola/domain` | Artifact, Command, Detection, … | scrape lola, graph ingest lola |
| `pkg/ds/domain` | Resource (Caldera, Sigma, …) | scrape ds, pipeline NED, graph ingest ds |
| `pkg/sbom/domain` | AdvisoryRef (OSV/GHSA) | scrape sbom, pipeline NED, graph ingest sbom |
| `pkg/nuclei/domain` | Template | scrape nuclei, pipeline NED, graph ingest nuclei |
| `pkg/coderules/domain` | RuleFile (Semgrep, CodeQL) | scrape coderules, pipeline NED, graph ingest coderules |
| `pkg/ti/validate` | Pure validation (type, empty fields) | `pkg/ti/normalize` |
| `pkg/ti/normalize` | NED normalization (IOC canonical form) | **pipeline only** — not graph ingest |
| `pkg/ti/ids` | Stable actor/report/IOC ids for dedup | pipeline NED (`normalize` re-exports) |
| `pkg/harvest` | Scrape → NATS envelope | scrape, pipeline |
| `pkg/commit` | Pipeline → graph ingest envelope | pipeline, graph |
| `pkg/engage/contract` | Engage HTTP/MCP DTOs | engage serve |
| `pkg/engage/domain/report` | Finding, Severity | engage serve |
| `pkg/engage/domain/job` | Job, Status | engage serve |
| `pkg/engage/domain/tool` | Tool spec, catalog metadata | engage serve |
| `pkg/engage/domain/target` | Scan subject (host, URL, CIDR) | engage serve (guard, workflows) |
| `pkg/engage/events` | Finding events wire | engage, pipeline, graph ingest |
| `pkg/playbook/domain` | `SkillMeta`, `IndexFile`, `SkillDetail` | Generated index + veil-api/MCP read (not ingest STIX) |
| `pkg/playbook/index` | Load/search `docs/skills-index/cyber-skills.json` | `knowledge/serve` playbook usecase |
| `pkg/playbook/corpus/mappings` | MITRE Navigator layer, NIST CSF, OWASP MD (committed) | `pkg/playbook/framework`, veil-api framework routes |
| `pkg/playbook/framework` | Parse Navigator layer JSON | HTTP/MCP `playbook_framework` |
| `pkg/playbook/procedure` | Structured ProcedureSpec; `procedures-index.json` | veil-api `/procedure`, MCP `playbook_procedure` |
| `pkg/playbook/cataloglink` | Map skill tool mentions → engage catalog names | `playbook_recommend_tools`, engage boost |

## Knowledge serve read DTOs (layer-local, not in pkg)

These types stay in the knowledge module because they are Neo4j projections and HTTP/MCP response shapes, not shared ingest entities. Engage reads them via HTTP veil-api only (no cross-layer Go import).

| Location | Types | Role |
|----------|-------|------|
| `knowledge/connector/query` | `Node`, `Edge`, `Graph`, `KindCount`, `CategoryMeta`, `EngageFindingContext`, `EngageTargetContext` | Cypher read projections for API/MCP |
| `knowledge/serve/internal/domain` | `ErrNodeNotFound` | HTTP 404 mapping for missing nodes |
| `knowledge/serve/internal/usecase` | `TargetGraphOpts`, `TargetGraphState` | Aggregated target read model over category search + engage subgraph |

Do not duplicate `pkg/ti/domain` entities here; graph nodes expose label/props maps for transport.

## Layer adapters (not in pkg)

| Layer | Path pattern | Responsibility |
|-------|--------------|----------------|
| Scrape | `discovery/harvest/internal/sources/<src>/` | Fetch, parse → `harvest.Envelope` (uses `pkg/*/domain`) |
| Pipeline | `pipeline/ned/internal/sources/<src>/` | Transform → normalize → `commit.Envelope` |
| Graph ingest | `knowledge/ingest/internal/sources/<src>/` | Apply commit → Neo4j (uses `pkg/*/domain`) |
| Engage | `engage/serve/internal/security`, usecase adapters | Target guard, tool runners (domain types in `pkg/engage/domain`) |

## TI flow

```text
scrape (raw IOC in harvest payload)
  → pipeline NED (pkg/ti/normalize + pkg/ti/ids)
  → commit envelope (normalized payload)
  → graph ingest (MERGE, no re-normalize)
```

## Deprecations

- NED normalization imports **`pkg/ti/normalize`** only (removed `pipeline/pkg/ti/normalize` forwarder, 2026-05 cleanup).
