# Client dependencies for Engage (host-native tools)

Engage runs catalog tools as **host-native subprocesses** on the same machine as the MCP server process. For where that process lives in your setup (Cursor, HTTP MCP, etc.), see [engage-mcp-topology.md](engage-mcp-topology.md).

**Veil does not auto-install** the binaries below. Install them yourself on the host that runs the Engage MCP server so they appear on `PATH` (same expectation as upstream HexStrike’s “Install Security Tools” flow). For Linux multi-distro automation hints, see [engage-install-linux.md](engage-install-linux.md).

Upstream checklist (full lists and install hints): [refs/hexstrike-ai-master/README.md](../refs/hexstrike-ai-master/README.md) — section **Install Security Tools** (Core, Cloud, Browser Agent).

## Mapping: categories → examples → checks

Representative binaries are **paraphrased** from the upstream README; treat the linked file as the source of truth for complete names and install steps.

### Core — network and reconnaissance

| Tool group / category | Example binaries (from upstream Core list) | Suggested check |
|-----------------------|---------------------------------------------|-----------------|
| Port and service scanning | `nmap`, `masscan`, `rustscan` | `command -v nmap` or `nmap --version` |
| DNS / subdomain / OSINT | `amass`, `subfinder`, `fierce`, `dnsenum`, `theharvester` | `command -v subfinder` or `subfinder -version` |
| Vulnerability templates / crawling | `nuclei`, `katana`, `httpx` | `command -v nuclei` or `nuclei -version` |
| AD / SMB-style enumeration | `netexec`, `enum4linux-ng`, `responder` | `command -v netexec` or `netexec --version` |
| Broad automation | `autorecon` | `command -v autorecon` |

### Core — web application security

| Tool group / category | Example binaries (from upstream Core list) | Suggested check |
|-----------------------|---------------------------------------------|-----------------|
| Content discovery / fuzzing | `gobuster`, `feroxbuster`, `ffuf`, `dirsearch`, `dirb` | `command -v ffuf` or `ffuf -V` |
| Parameter / WAF testing | `arjun`, `paramspider`, `dalfox`, `wafw00f` | `command -v sqlmap` or `sqlmap --version` |
| Classic web scanners | `nikto`, `wpscan` | `command -v nikto` |

### Core — passwords, auth, and secrets

| Tool group / category | Example binaries (from upstream Core list) | Suggested check |
|-----------------------|---------------------------------------------|-----------------|
| Cracking / guessing | `hydra`, `john`, `hashcat`, `medusa`, `patator` | `command -v hashcat` or `hashcat --version` |
| Windows remote / hashes | `crackmapexec`, `evil-winrm`, `hash-identifier` | `command -v evil-winrm` |

### Core — binary analysis and forensics

| Tool group / category | Example binaries (from upstream Core list) | Suggested check |
|-----------------------|---------------------------------------------|-----------------|
| Debug / RE | `gdb`, `radare2`, `binwalk`, `ghidra`, `checksec`, `strings`, `objdump` | `command -v r2` or `r2 -v` |
| Memory / media | `volatility3`, `foremost`, `steghide`, `exiftool` | `command -v exiftool` or `exiftool -ver` |

### Cloud and container security

| Tool group / category | Example binaries (from upstream Cloud list) | Suggested check |
|-----------------------|----------------------------------------------|-----------------|
| Multi-cloud auditing | `prowler`, `scout-suite` | `command -v prowler` or `prowler -v` |
| Image / workload scanning | `trivy` | `command -v trivy` or `trivy --version` |
| Kubernetes / Docker hardening | `kube-hunter`, `kube-bench`, `docker-bench-security` | `command -v kube-hunter` or `kube-hunter --help` |

### Browser agent (headless automation)

| Tool group / category | Example binaries / packages (from upstream Browser section) | Suggested check |
|-----------------------|--------------------------------------------------------------|-----------------|
| Chromium stack | `chromium-browser`, `chromium-chromedriver` (distro packages) or Google Chrome | `command -v chromium` or `chromium --version`; `command -v chromedriver` or `chromedriver --version` |
| Google Chrome (optional path) | `google-chrome-stable` when installed from vendor repo | `command -v google-chrome-stable` or `google-chrome-stable --version` |

Use the upstream README for exact `apt` / vendor install lines; this document only records **what must exist on the MCP host** for parity with that contract.

**Generated coverage matrix (158 tools):** `make engage-tool-install-coverage` → [engage-tool-install-coverage.md](engage-tool-install-coverage.md). **Profiles and installer:** [engage-install-linux.md](engage-install-linux.md).
