# Knowledge layer

Part of the four-layer Veil stack — [README.md](../README.md#architecture). Architecture rules: [docs/agents/coding-style.md](../docs/agents/coding-style.md).

Consumes `ingest.>`, MERGE into Neo4j; HTTP API and MCP read Bolt. Ingest includes TI/vuln/lola/ds/SBOM sources plus optional **engage** (`SourceEngage` → `EngageToolRun` / `EngageFinding` in [ingest/internal/sources/engage/](ingest/internal/sources/engage/)). **Tool execution** is a separate layer: [engage/README.md](../engage/README.md) (`veil-engage` MCP), which may call this API read-only via JWT and optionally publish scan metadata on `engage.events.>`.

| Module | Path | Role |
|--------|------|------|
| **ingest** | [ingest/](ingest/) | JetStream pull consumer (`ingest_worker`) → Neo4j |
| **serve** | [serve/](serve/) | HTTP API (`api`) and MCP (`mcp`) — [docs/agents/mcp-agents.md](../docs/agents/mcp-agents.md) |
| **connector** | [connector/](connector/) | Shared Bolt driver and categorical queries |

- **Wire types:** [pkg/commit/](../pkg/commit/)
- **Build / test:**

```bash
make test-knowledge
make test-knowledge-serve    # knowledge/serve only, -race
```

```bash
cd knowledge/ingest && go build -o bin/ingest_worker ./cmd/ingest_worker
cd knowledge/serve && go build -o bin/api ./cmd/api && go build -o bin/mcp ./cmd/mcp
```

- **Deploy (dev):** [deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml) — Neo4j `7474/7687`, API `8090`, MCP HTTP `8091` (`--profile mcp`)
- **Graph read smoke:** `make test-knowledge-read-smoke` — [compose.graph-read.yml](../deploy/knowledge/compose.graph-read.yml) (no ingest/NATS)
- **Secure prod:** [compose.secure.yml](../deploy/knowledge/compose.secure.yml) + [docs/deploy/deploy-secure.md](../docs/deploy/deploy-secure.md)

## serve layout

```
serve/
  cmd/{api,mcp}/
  internal/
    auth/              # JWT (Keycloak JWKS), RBAC, MCP authorize
    components/        # DI: Neo4j, ReadUsecase, auth stack
    config/            # env: API, MCP HTTP, security hardening
    transport/
      httpserver/      # REST /v1/*
      mcpserver/       # stdio + Streamable HTTP
      securityhttp/    # headers, body limits, CORS allowlist
    usecase/           # ReadUsecase (shared API + MCP)
    storage/neo4j/
```

Optional auth: [docs/deploy/auth-keycloak.md](../docs/deploy/auth-keycloak.md). Distroless images with `/api healthcheck` and `/mcp healthcheck` for container probes.

## ingest layout

```
ingest/
  cmd/ingest_worker/     # thin main
  internal/
    components/          # DI: Neo4j writers + domain appliers
    ingest/              # NATS pull loop
    sources/{ti,vuln,lola,ds}/
    appsec/{sbom,coderules,nuclei}/
```
