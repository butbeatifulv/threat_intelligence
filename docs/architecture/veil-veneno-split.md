# ADR: veil (knowledge) + veneno (pentest) split

**Status:** Accepted (in progress)  
**Date:** 2026-06-23

## Context

Veil monorepo contains four isolated Go layers: `discovery/`, `pipeline/`, `knowledge/`, `engage/`. The engage layer is pentest execution (HexStrike successor). Knowledge is TI graph ingest + read API/MCP.

## Decision

1. **veil** remains the knowledge/TI repository: discovery → pipeline → knowledge, `veil-api`, `veil-mcp`, graph pack, playbook corpus.
2. **veneno** is a new repository for pentest execution: `engage/`, tool catalog, `veneno-api`, `veneno-mcp`, HexStrike extract tooling.
3. Integration **only** via frozen contracts:
   - HTTP: veneno → veil `GET /v1/*` (veilgraph client)
   - NATS: veneno → `engage.events.*` → veil `pipeline/engage-events` → `ingest.engage.*`
4. `pipeline/engage-events` and knowledge ingest `engage` source **stay in veil**.

## Non-goals (this split)

- cxado shared Go modules (`shared/pkg`) — future phase
- `pkg/domain` meta-layer migration — future phase
- cys-agi MCP wiring — separate PR after veneno submodule
- Renaming internal `engage/` paths inside veneno — optional follow-up wave

## Consequences

- `pkg/engage/events` wire types duplicated in veneno; veil `pipeline/` keeps import for bridge consumer
- `pkg/engage/hostnorm` stays in **veil** (used by `knowledge/connector/query`)
- Unified edge P12 in veil proxies `/api/*` and `/mcp/engage` to veneno upstreams

## References

- [split-r0-inventory.md](../development/split-r0-inventory.md)
- [ingest-contract.md](../contracts/ingest-contract.md)
- [cxado plan](../../../.cursor/plans/veil_veneno_split_2ab85428.plan.md) (meta-repo)
