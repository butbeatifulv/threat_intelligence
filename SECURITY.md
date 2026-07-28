# Security

## Reporting a vulnerability

If you believe you have found a security vulnerability in this repository (for example in the HTTP API, MCP server, or a dependency choice), please report it responsibly:

- Prefer **GitHub Security Advisories** for this repository (private disclosure), if enabled for the project.
- Otherwise, open a **private** issue or contact the maintainers through a channel they publish on the repository or organization profile.

Please do **not** file public issues with exploit details until a fix is available.

## Scope

This project runs **Neo4j**, **Go services**, and optional **scrapers** with network access to third-party APIs. Harden deployments with secrets management, non-default passwords, network policies, and regular image updates—not only application patches.

## Authentication (graph read)

The HTTP API and MCP (stdio and Streamable HTTP) support **optional** JWT validation via **Keycloak** (`AUTH_ENABLED`, default off). When enabled, configure `KEYCLOAK_ISSUER`, protect client secrets, and use RBAC roles (`veil-reader`, `veil-admin`). Do not commit tokens or client secrets. See [docs/deploy/auth-keycloak.md](docs/deploy/auth-keycloak.md).

For production-style deployments (TLS edge, distroless images, fail-closed auth, no published Neo4j ports), use the secure compose overlay and checklist in [docs/deploy/deploy-secure.md](docs/deploy/deploy-secure.md). Set `VEIL_REQUIRE_AUTH=1` so services refuse to start without Keycloak when auth is mandatory.

## Engage (active tooling layer)

The **engage** layer executes third-party security tools. For secured infrastructure (active threat countermeasures), use [docs/engage/engage-hardening.md](docs/engage/engage-hardening.md): Docker-isolated runner, deny raw `/api/command`, auth required, and `make test-engage-hardening` (safe self-test — no host exploitation).

## Supported versions

Security fixes are applied on a best-effort basis on the default development branch. There is no separate LTS branch unless documented explicitly in the future.
