# ingest (graph write path)

NATS JetStream consumer for `ingest.>` → Neo4j MERGE.

- **Binary:** `cmd/ingest_worker` (compose service name unchanged: `ingest_worker`)
- **Wire types:** [pkg/commit](../../pkg/commit/)
- **Domain writers:** `internal/sources/{ti,vuln,lola,ds}/`
- **AppSec storage:** `internal/appsec/{sbom,coderules,nuclei}/`
- **Deploy:** [deploy/knowledge/compose.yml](../../deploy/knowledge/compose.yml)

```bash
cd knowledge/ingest && go build -o bin/ingest_worker ./cmd/ingest_worker
```
