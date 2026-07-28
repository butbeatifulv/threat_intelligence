# GAIA evaluation (Veil)

**Primary reference:** [GAIA: A Benchmark for General AI Assistants](https://arxiv.org/abs/2311.12983) (arXiv:2311.12983).

Veil does **not** require a Hugging Face token for CI. Use:

| Split | Path | Purpose |
|-------|------|---------|
| Pilot (synthetic) | `fixtures/pilot/` | Harness smoke |
| Paper examples (public) | `fixtures/paper-examples/` | Scorer + level format (Fig. 1) |
| Full 466 tasks | HF gated / manual JSONL | Optional local only |

Docs: [docs/agents/agent-evaluation-gaia.md](../../docs/agents/agent-evaluation-gaia.md).
