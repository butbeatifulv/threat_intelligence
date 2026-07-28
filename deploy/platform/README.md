# Veil platform deploy (P12)

Unified access layer: one TLS nginx edge for graph (`api`, `mcp`) and engage (`engage-api`, `engage-mcp`) backends.

## Layout

| Path | Role |
|------|------|
| [nginx/veil.conf](nginx/veil.conf) | Route map for unified edge |
| [nginx/upstreams.conf](nginx/upstreams.conf) | Docker service upstreams |
| [compose.edge.yml](compose.edge.yml) | Compose overlay (unpublishes layer ports, adds `veil-edge`) |
| [compose.edge-hostless.yml](compose.edge-hostless.yml) | CI/smoke: `ports: !reset []` when merge appends host binds |
| [docker/nginx.Dockerfile](docker/nginx.Dockerfile) | `veil-edge` image |
| [../stacks/unified-edge.yml](../stacks/unified-edge.yml) | Full stack preset |

## Route map

| Path | Backend | Port (container) |
|------|---------|------------------|
| `/health` | `api` | 8090 |
| `/v1/` | `api` | 8090 |
| `/api/` | `engage-api` | 8890 |
| `/mcp/graph/` | `mcp` → `/mcp/` | 8091 |
| `/mcp/engage/` | `engage-mcp` → `/mcp/` | 8892 |

Host ingress: `${VEIL_EDGE_HTTPS_PORT:-443}` (HTTPS only).

## Bring up (unified-edge stack)

```bash
# Dev TLS (once)
mkdir -p deploy/platform/nginx/certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deploy/platform/nginx/certs/tls.key \
  -out deploy/platform/nginx/certs/tls.crt \
  -subj '/CN=localhost'

# From repo root — same file chain as deploy/stacks/unified-edge.yml
docker compose \
  -f deploy/discovery/compose.yml \
  -f deploy/pipeline/compose.yml \
  -f deploy/knowledge/compose.yml \
  -f deploy/knowledge/compose.neo4j-publish.yml \
  -f deploy/engage/compose.yml \
  -f deploy/engage/compose.veil-stack.yml \
  -f deploy/platform/compose.edge.yml \
  --profile mcp \
  --env-file deploy/profiles/smoke-minimal.env \
  up -d --build
```

## Automated smoke (P12i)

```bash
make test-platform-unified-edge
# or: ./scripts/test/smoke-unified-edge.sh [--up] [--down]
```

Skip in CI locally: `SMOKE_SKIP_UNIFIED_EDGE=1`.

## Curl smoke (through veil-edge)

Assume `VEIL_EDGE_HTTPS_PORT=443` and dev cert from above (`-k` skips verify).

```bash
BASE="https://127.0.0.1:${VEIL_EDGE_HTTPS_PORT:-443}"

curl -skS "$BASE/health"
curl -skS "$BASE/v1/categories" | head -c 200; echo
curl -skS "$BASE/api/tools" | head -c 200; echo

# MCP Streamable HTTP (POST initialize); graph path prefix strips to backend /mcp/
curl -skS -X POST "$BASE/mcp/graph/" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}'

curl -skS -X POST "$BASE/mcp/engage/" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}'
```

Direct backend ports are not published when `compose.edge.yml` is applied; use the edge URLs only.

## Related

- P12 master plan: [.cursor/plans/archive/veil_platform_p12_unified_access.plan.md](../../.cursor/plans/archive/veil_platform_p12_unified_access.plan.md)
- Layer-specific nginx (legacy): `deploy/knowledge/compose.secure.yml`, `deploy/engage/compose.secure.yml`
