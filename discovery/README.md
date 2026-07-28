# Discovery layer (harvest)

Part of the four-layer Veil stack — [README.md](../README.md#architecture). Architecture rules: [docs/agents/coding-style.md](../docs/agents/coding-style.md).

Fetches third-party feeds, records fetches in the Vitess/MySQL **ledger**, and publishes **`harvest`** envelopes to NATS **`scrape.>`**.

| Module | Path | Role |
|--------|------|------|
| **connector** | [connector/](connector/) | JetStream publish + stream ensure |
| **harvest** | [harvest/](harvest/) | Batch worker (`scrape_worker`) + sources |
| **proxybroker** | [proxybroker/](proxybroker/) | HTTP proxy pool sidecar |

- **Wire types:** [pkg/harvest/](../pkg/harvest/)
- **Build:** `make test-discovery` or:

```bash
cd discovery/harvest && go build -o bin/scrape_worker ./cmd/scrape_worker
```

- **Deploy:** [deploy/discovery/compose.yml](../deploy/discovery/compose.yml)

## harvest layout

```
harvest/
  cmd/scrape_worker/     # thin main + blank-import sources
  internal/
    factory/             # registry, Run, subjects
    feeds/ ledger/         # shared fetch + crawl_resource
    sources/{ti,vuln,lola,ds,sbom,coderules,nuclei}/
```

## Sources (`SCRAPE_SOURCES`)

Default (compose): `ds,vuln,lola,ti,sbom,coderules,nuclei`

| Source | Publishes | Notable env |
|--------|-----------|-------------|
| `vuln` | NVD pages, Exploit-DB, optional Vulners/MSF | `NVD_MAX_PAGES`, `NVD_API_KEY` |
| `lola` | LOLBAS, GTFOBins, LOFTS, MITRE STIX | `LOLA_MITRE_MAX_TECHNIQUES`, `LOFTS_URL` |
| `ds` | Sigma, YARA, Atomic Red Team, Caldera | `DS_MAX_SIGMA`, `DS_MAX_YARA` |
| `ti` | KEV, ThreatFox, MalwareBazaar, Feodo, JSONL | `TI_FEEDS`, `TI_JSONL_FILE` |
| `sbom` | OSV, GHSA | `SBOM_SOURCES`, `SBOM_MAX_CVES` |
| `coderules` | Semgrep, CodeQL, CWE catalog | `CODERULES_MAX_SEMGREP` |
| `nuclei` | Nuclei templates | `NUCLEI_MAX` |

Ledger: `VITESS_DSN`, `SCRAPE_MIN_REFETCH_AFTER`, `SCRAPE_FORCE_REFETCH=1` for full refetch.

## Local run

```bash
export NATS_URL=nats://localhost:4222
export SCRAPE_SOURCES=vuln
export NVD_MAX_PAGES=1
cd discovery/harvest && go run ./cmd/scrape_worker
```

Full stack: `./scripts/ops/compose-up-full.sh` from repo root.

## Further reading

- [docs/architecture/threatintel-runtime.md](../docs/architecture/threatintel-runtime.md)
- [docs/contracts/ingest-contract.md](../docs/contracts/ingest-contract.md)
- [pipeline/README.md](../pipeline/README.md) — downstream NED
