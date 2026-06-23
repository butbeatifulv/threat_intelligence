# Engage client-native tools on Linux (multi-distro)

Engage runs catalog tools as **host subprocesses** (see [engage-mcp-topology.md](engage-mcp-topology.md)). This runbook covers installing common CLIs with the **system package manager** and verifying them with preflight.

## Prerequisites

- **Python 3** with **PyYAML** (for reading `scripts/ops/engage-tools-packages.yaml`):
  - Debian/Ubuntu: `sudo apt-get install -y python3-yaml`
  - Fedora: `sudo dnf install -y python3-pyyaml`
  - Arch: `sudo pacman -S --needed python-yaml`

## Package map

Authoritative mapping of profile → tool → distro packages:

- [`../scripts/ops/engage-tools-packages.yaml`](../scripts/ops/engage-tools-packages.yaml)
- [`../scripts/ops/engage-tools-sources.yaml`](../scripts/ops/engage-tools-sources.yaml) (Kali/pkg tracker/upstream provenance + fallback methods)

Profiles: `minimal`, `recommended`, `full` (see file). Some tools have **empty** package lists on certain distros (for example `masscan` on Alpine); install those manually or skip them in that environment.
For HexStrike parity there is a dedicated `core47` profile based on [`refs/hexstrike-ai-master/README.md`](../refs/hexstrike-ai-master/README.md).

## Plan-only vs install

From the repo root:

```bash
# Show the command the installer would run (no packages installed)
make engage-install-plan

# Or: ./scripts/ops/install-engage-host-tools.sh --plan --profile minimal
```

To **install** packages (uses `sudo`):

```bash
make engage-install-host-tools

# Or pick profile explicitly:
ENGAGE_INSTALL_PROFILE=minimal ./scripts/ops/install-engage-host-tools.sh --yes
```

If your distro repositories miss some tools, choose an explicit policy:

```bash
# upstream fallback (go/cargo/pipx/gem from source registry)
make engage-install-fallback

# Kali fallback for Debian/Ubuntu (opt-in, pinned allowlist)
make engage-install-kali-fallback

# Explicit policy form:
./scripts/ops/install-engage-host-tools.sh --yes --profile recommended --policy upstream-fallback
./scripts/ops/install-engage-host-tools.sh --yes --profile recommended --policy kali-fallback
./scripts/ops/install-engage-host-tools.sh --yes --profile recommended --policy full-auto
./scripts/ops/install-engage-host-tools.sh --yes --profile core47 --policy full-auto
```

Override the YAML path with `ENGAGE_TOOLS_PACKAGES_YAML` if you maintain a forked map.

## Preflight

```bash
./scripts/engage/preflight-client-tools.sh
./scripts/engage/preflight-client-tools.sh --profile minimal
./scripts/engage/preflight-client-tools.sh --profile full --json
./scripts/engage/preflight-client-tools.sh --profile recommended --emit-missing
./scripts/engage/preflight-client-tools.sh --profile recommended --emit-install-plan --policy full-auto
./scripts/engage/preflight-client-tools.sh --profile core47 --json
```

Environment: `ENGAGE_PREFLIGHT_PROFILE`, `ENGAGE_TOOLS_PACKAGES_YAML`, `ENGAGE_TOOLS_SOURCES_YAML`, `ENGAGE_INSTALL_POLICY`.

## Coverage artifacts (158-tool track)

```bash
make engage-tool-source-map
make engage-tool-install-coverage
```

Artifacts:

- [`engage-tools-sources.yaml`](../scripts/ops/engage-tools-sources.yaml) — source provenance + fallback methods
- [`engage-tool-install-coverage.md`](engage-tool-install-coverage.md) — per-tool status for Ubuntu/Debian repo, Kali fallback, upstream fallback

## Core47 one-shot install (client quick path)

```bash
# Optional: disable flaky third-party apt repos first if update hangs
./scripts/ops/install-engage-host-tools.sh --yes --profile core47 --policy full-auto
./scripts/engage/preflight-client-tools.sh --profile core47 --json
export ENGAGE_VICTIM_URL=http://127.0.0.1:8891
make test-engage-red-blue
```

## Red-vs-blue lab (optional)

After `engage-api` is running as **victim** on `:8891`, run the harness:

```bash
export ENGAGE_VICTIM_URL=http://127.0.0.1:8891
make test-engage-red-blue
```

Details and legal scope: [engage-red-blue-lab.md](engage-red-blue-lab.md).

## See also

- [engage-client-dependencies.md](engage-client-dependencies.md) — what must exist on `PATH`
- [../engage/README.md](../engage/README.md) — dev compose vs client-native
- [../AGENTS.md](../AGENTS.md) — agent quick paths (recommended + core47)
