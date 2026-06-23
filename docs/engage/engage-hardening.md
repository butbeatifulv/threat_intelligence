# Engage layer hardening (active defense infrastructure)

Veil **engage** runs offensive tooling inside a **protected perimeter** for active threat countermeasures. Hardening assumes attackers may probe the API/MCP surface; the host and graph must not become pivot points.

**Framework alignment:** Jet [JCSF](../refs/Jet-Container-Security-Framework-main/), [DAF](../refs/DevSecOps-Assessment-Framework-main/) (+ MLSO for agentic), OWASP GenAI agentic/red-team landscapes — see [external-security-frameworks.md](../external/external-security-frameworks.md) and [engage-agentic-threats.md](engage-agentic-threats.md). Enforced catalog: [deploy/security/veil-controls.yaml](../deploy/security/veil-controls.yaml).

## Threat model

| Threat | Impact | Mitigations in repo |
|--------|--------|---------------------|
| Unauthenticated tool execution | Full compromise of runner/network | `AUTH_ENABLED`, `VEIL_REQUIRE_AUTH`, Keycloak RBAC |
| Raw shell via `/api/command` | RCE on API or runner host | Catalog allowlist; `ENGAGE_DENY_RAW_COMMAND=1`; shell metachar block |
| Local runner in prod | Tools execute on API host | Prefer isolated runner: `ENGAGE_EXECUTION_PROFILE=docker-exec` with `ENGAGE_RUNNER_MODE=docker` + `engage-runner`; default `client-native` forbids docker runner in-process |
| Path traversal on file APIs | Read/write host files | Chrooted `files.Manager` under `ENGAGE_FILES_DIR` |
| MCP/HTTP abuse | DoS, header injection | `securityhttp` limits, timeouts, CSP, no `Server` leak |
| Weak credentials | Graph/tool data exfil | Rotate `NEO4J_PASS`; secrets manager in prod |
| Supply chain / catalog | Malicious tool defs | `tools.live.yaml` review; enable-by-category |

**Out of scope for automated self-test:** exploitation of third-party binaries (nmap, nuclei, …). Self-test validates **configuration and guard code**, not CVEs inside upstream tools.

## Production profile (minimum)

Use [deploy/profiles/secure-engage.env](../deploy/profiles/secure-engage.env) with [compose.secure.yml](../deploy/engage/compose.secure.yml):

| Variable | Value | Purpose |
|----------|-------|---------|
| `ENGAGE_ENV` | `prod` | Strict errors, HSTS, strict child env |
| `ENGAGE_DENY_RAW_COMMAND` | `1` | Disable raw `/api/command` |
| `ENGAGE_STRICT_ENV` | `1` | Minimal env passed to subprocesses |
| `ENGAGE_EXECUTION_PROFILE` | `docker-exec` | Required with docker runner (`client-native` forbids `ENGAGE_RUNNER_MODE=docker` in API) |
| `ENGAGE_RUNNER_MODE` | `docker` | Never `local` when using docker runner in prod |
| `ENGAGE_RUNNER_CONTAINER` | `engage-runner` | Toolbox isolation |
| `VEIL_REQUIRE_AUTH` | `1` | Fail closed without auth |
| `ENGAGE_MCP_HTTP_AUTH_STRICT` | `1` | Bearer on MCP HTTP |
| `ENGAGE_TARGET_GUARD` | `block` | Block metadata/RFC1918/loopback tool targets |

Overlay: TLS nginx only on host ([docs/deploy/deploy-secure.md](../deploy/deploy-secure.md)).

## Self-test (safe on your machine)

Does **not** attack your OS or scan the internet:

```bash
make test-engage-hardening
```

Includes:

1. `go test` — `engage/serve/internal/security` (config + injection guards)
2. `scripts/engage/hardening-compose-audit.py` — static compose/profile checks
3. Optional Docker — `smoke-engage-secure.sh` (ephemeral project, torn down on exit)

Strict mode (fail on high findings in current env):

```bash
ENGAGE_HARDENING_SELFTEST_STRICT=1 make test-engage-hardening
```

API startup fail-closed (staging/prod deploy):

```bash
ENGAGE_HARDENING_FAIL_ON=high ENGAGE_ENV=prod ...
```

## CI

| Workflow | When |
|----------|------|
| [engage.yml](../.github/workflows/engage.yml) | PR: hardening unit tests |
| [engage-secure.yml](../.github/workflows/engage-secure.yml) | Nightly: self-test + secure smoke |

## Operator checklist (secured infra)

- [ ] Runner in Docker, not local mode
- [ ] Raw command API disabled
- [ ] Auth required; MCP strict in prod
- [ ] TLS termination; only 443 (or 8443) on host
- [ ] NetworkPolicy / firewall: engage-api not exposed to untrusted networks
- [ ] Central audit logs (`ENGAGE_AUDIT_*`, webhook with secret)
- [ ] Regular image digest pins and runner image rebuilds

## External frameworks (JCSF, DAF, OWASP Agentic)

- [external-security-frameworks.md](../external/external-security-frameworks.md) — reference index (`.external/` read-only)
- [engage-agentic-threats.md](engage-agentic-threats.md) — agent/MCP threats ↔ controls
- [deploy/security/veil-controls.yaml](../deploy/security/veil-controls.yaml) — `python3 scripts/engage/hardening-framework-audit.py`

See also [SECURITY.md](../SECURITY.md), [engage-runtime.md](engage-runtime.md), [deploy-platform-hybrid.md](../deploy/deploy-platform-hybrid.md).
