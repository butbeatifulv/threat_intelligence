# harvest (Discovery acquisition path)

Batch worker: fetch external feeds, ledger dedup, publish `harvest` to `scrape.>`.

- **Binary:** `cmd/scrape_worker` (compose service name unchanged: `scrape_worker`)
- **Fetch / ledger:** `internal/feeds`, `internal/ledger`
- **Sources:** `internal/sources/{ti,vuln,lola,ds,sbom,coderules,nuclei}`
- **NATS publish:** [discovery/connector](../connector/)
- **Deploy:** [deploy/discovery/compose.yml](../../deploy/discovery/compose.yml)

```bash
cd discovery/harvest && go build -o bin/scrape_worker ./cmd/scrape_worker
```
