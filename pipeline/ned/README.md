# ned (pipeline transform path)

NATS JetStream consumer for `scrape.>` → NED → `ingest.>`.

- **Binary:** `cmd/pipeline_worker` (compose service name unchanged: `pipeline_worker`)
- **NED:** Normalize (`pkg/ti/normalize`, AppSec YAML parse) → Enrich (NVD CWE/CPE via `pipeline/pkg/nvd/parse`) → Dedup (`commit` idempotency keys + JetStream `Nats-Msg-Id`)
- **Wire types:** [pkg/harvest](../../pkg/harvest/), [pkg/commit](../../pkg/commit/)
- **NATS publish:** [pipeline/connector](../connector/)
- **Deploy:** [deploy/pipeline/compose.yml](../../deploy/pipeline/compose.yml)

```bash
cd pipeline/ned && go build -o bin/pipeline_worker ./cmd/pipeline_worker
```
