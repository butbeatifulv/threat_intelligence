# Security Gate (`.github/workflows/security-gate.yml`)

DevSecOps gate for veil's Go monorepo: gosec, govulncheck and golangci-lint
run as a per-module matrix (21 independent `go.mod`s under `pkg/`,
`discovery/`, `pipeline/`, `knowledge/`, `platform/`, discovered at runtime
by `discover-modules` rather than hardcoded), plus repo-wide CodeQL (Go) and
gitleaks. All five are blocking — there is no warn-only tier in this gate,
because every finding it started with has been fixed rather than deferred
(see below). `.golangci.yml` at the repo root configures the one legitimate
exception: `errcheck` doesn't flag the `defer resp.Body.Close()` /
`nc.Drain()` / `sub.Unsubscribe()` idiom used consistently across this repo.

## What this gate found and fixed on its first real run

### `platform.yml`'s `unit` job had been red on every push to main since 2026-06-24

Root cause: commit `53ee58a` ("Split engage layer to veneno") deleted
`pkg/exec` (the sandbox executor, now in veneno), but `discovery/pkg/go.mod`
kept a `require`+`replace` pointing at it. The only thing that imported it
was `discovery/pkg/execfetch/git_discoveryexec.go`, an experimental
build-tag-gated spike (`//go:build discoveryexec`) that could never compile
without the deleted package. Go's module-graph resolution needs to resolve
`require`/`replace` directives independent of build tags, so this broke
`go build`/`go test` for the whole `discovery` workspace (`discovery/pkg`,
`discovery/browser`, `discovery/connector`, and anything importing
`pkg/natsjet`) even though the dead file itself never compiled in a normal
build. Fixed by deleting the orphaned spike and the dangling require/replace.
Confirmed via `gh run list --workflow=platform.yml`: 5 consecutive failures
going back 3+ weeks, all the same root cause.

### Two real, copy-paste compile bugs in `discovery/harvest`

- `internal/sources/vuln/internal/usecase/scrape.go`: `"log/slog"` imported
  twice — duplicate-import compile error.
- `internal/sources/vuln/internal/usecase/exploits.go`:
  `ingestVulnersSearch` (single `error` return) had a stray
  `return feeds.FetchResult{}, ctx.Err()` copy-pasted from
  `fetchBytesExploit` (which does return `(feeds.FetchResult, error)`) —
  "too many return values" compile error.

### `pkg/engage/go.mod` was missing a `replace` for `pkg/observability`

`pkg/engage` transitively imports `pkg/observability` via `pkg/natsjet`, but
had no `replace` for it (every other module that needs it does). Go tried to
fetch a real tagged release of a module that's never published, and failed.
Added the missing `replace ... => ../observability`, matching the pattern
already used in `knowledge/ingest`, `pipeline/ned`, etc.

### Go toolchain patch version — 23 real stdlib CVEs

All 21 modules pinned `go 1.25.0` with no `toolchain` directive, so
`GOTOOLCHAIN=auto` used exactly that patch. `govulncheck` found 23 CVEs in
`crypto/x509`, `net/http`, `crypto/tls`, `encoding/asn1`, `encoding/pem`,
`net/url` — all already fixed in later 1.25.x patches. Fixed by pinning
`toolchain go1.25.12` (current latest 1.25.x) everywhere via `go mod edit
-toolchain=go1.25.12` / `go work edit -toolchain=go1.25.12`, leaving the
`go 1.25.0` language-version floor untouched.

### Two real third-party CVEs

- `go.opentelemetry.io/otel/sdk@v1.35.0` — GO-2026-4394 (arbitrary code
  execution via PATH hijacking), fixed in v1.40.0. Bumped in
  `pkg/observability` (the direct dependency) and in the 4 modules that
  pull it as a standalone module-proxy dependency rather than through the
  local workspace (`knowledge/ingest`, `pipeline/ned`, `knowledge/connector`,
  `pipeline/connector`).
- `golang.org/x/net@v0.51.0` (in `discovery/harvest`, reached via
  `html.Parse` in the LOFTS scraper) — GO-2026-5025, fixed in v0.55.0.

### 28 real gosec findings

- 14× G304 (file inclusion via variable) — reviewed each call site; all are
  operator/config-supplied paths (env vars, CLI args, or map lookups
  constrained to a committed index file's own keys), never
  attacker/network-controlled. Suppressed with `#nosec G304` plus a
  site-specific justification comment, not silenced.
- 1× G101 (`pkg/retrieval/config.go`) — flagged an env-var *name* constant
  (`"VEIL_EMBED_API_KEY"`), not a secret value. `#nosec G101` + reason.
- 3× G404 (weak RNG) — all cooldown/pacing jitter timing, not
  security-sensitive. `#nosec G404` + reason.
- 6× G104 (unchecked errors) — real fix: explicit `_ =` on best-effort
  cleanup calls, `strconv.Atoi` with a kept-default fallback instead of
  `fmt.Sscanf`, checked `os.Stderr.WriteString`.
- 2× G301 + 1× G306 (over-permissive file/dir modes) — real fix: `0o750`
  dirs, `0o600` files.
- 1× G112 (`platform/mcp-gateway`) — real fix: added
  `ReadHeaderTimeout` to the `http.Server`.

### 47 real golangci-lint findings

Most were the `defer x.Close()`/`Drain()`/`Unsubscribe()` idiom, now covered
by `.golangci.yml`'s `errcheck.exclude-functions`. The remainder were fixed
individually: a real dead-store bug in `pkg/retrieval/bleve.go` (a computed
`snippet` truncation was discarded and the untruncated `text` returned
instead — now wired through correctly), a deprecated `neo4j.Config` type
alias swapped for `config.Config` per the driver's own deprecation notice,
several safe mechanical simplifications (`S1011`, `QF1006`, `QF1001`), and
genuinely dead code removed (3 unused env consts, an unused `rpcError` type
alias + `toRPCError`/`rpcErr` helpers in `knowledge/serve`).

## Design notes

- **No reusable-workflow (`workflow_call`) layering.** egregore's
  equivalent gate hit a `startup_failure` (zero jobs, zero logs) because a
  called reusable workflow's job requested a `permissions:` scope its caller
  didn't hold at the top level. This workflow is a single flat file
  specifically to avoid that failure class.
- **Per-module matrix, not a single `./...`.** Go workspace mode doesn't let
  `./...` reach across module boundaries; each of the 21 `go.mod`s is
  scanned standalone (`GOWORK=off`), matching how this repo's own `Makefile`
  already runs tests per module.
- **gosec's own exit code is the gate**, verified empirically (0 = clean,
  1 = findings) — no SARIF-severity-inference layer needed, unlike CodeQL.
- **CodeQL's local SARIF has no severity metadata** (confirmed the same way
  egregore's did) — the `Gate check CodeQL` step fails on any finding
  (`COUNT > 0`), not severity-graded. Coarser, but never silently passes a
  real finding.
