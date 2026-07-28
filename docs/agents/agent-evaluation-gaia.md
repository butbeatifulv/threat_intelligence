# Agent evaluation — GAIA benchmark

**Canonical methodology:** [GAIA: A Benchmark for General AI Assistants](https://arxiv.org/pdf/2311.12983) (arXiv:2311.12983).

GAIA measures **general AI assistants** (reasoning, multi-modality, web browsing, tool use) with short, **unambiguous factual answers**. Veil uses it for **agent/orchestrator quality**, separate from **engage tool parity** and **security hardening**.

| Concern | Artifact | CI (no HF token) |
|---------|----------|------------------|
| Harness smoke | `eval/gaia/fixtures/pilot/` | `make test-agent-eval-pilot` |
| Paper-aligned format | `eval/gaia/fixtures/paper-examples/` | `make test-agent-eval-paper` |
| Full 466 questions | Local JSONL only | not in CI |
| Security tools | veneno catalog parity tests | veneno PRs |

## Hugging Face — optional, not required for Veil

The [HF dataset](https://huggingface.co/datasets/gaia-benchmark/GAIA) hosts the leaderboard splits. It is **gated** (not the same as API “rate limits”):

| Mechanism | What it means |
|-----------|----------------|
| **Gating** | You must log in, accept dataset terms (contact info / anti-leakage policy), then download. No token ⇒ no automated download. |
| **Rate limits** | HF Hub may throttle anonymous vs authenticated requests; irrelevant if you do not use HF at all. |
| **Contamination policy** | Do not commit validation/test rows or answers; do not publish private splits in crawlable form. |

**Veil default:** follow the **paper** (§3.1–3.4, §3.2 scoring, Figure 2 prompt). HF is an optional convenience for anyone who already has access.

### Without a token you can still

1. Run CI harness: `make test-agent-eval-pilot` and `make test-agent-eval-paper`.
2. Use the **system prompt** from the paper: `eval/gaia/paper/system-prompt.txt`.
3. Score any JSONL you create manually (`task_id`, `Question`, `Level`, `Final answer`) with `scripts/eval/gaia/score.py`.
4. Extend fixtures using the paper’s **question-design guidelines** (§3.4) — curated, unambiguous, single correct answer.

### With HF access (optional)

```bash
cp deploy/eval/gaia.env.example deploy/eval/gaia.env
# HF_TOKEN=...   # only if you accepted https://huggingface.co/datasets/gaia-benchmark/GAIA
pip install huggingface_hub datasets
python3 scripts/eval/gaia/load_hf.py   # writes eval/gaia/data/ (gitignored)
```

## Methodology (from the paper)

- **466 questions**, 3 **levels** (more steps / tools ⇒ harder). Human annotators ≈ **92%** vs best LLM+plugins ≈ **15–30%** on early tasks ([§1](https://arxiv.org/pdf/2311.12983)).
- **Developer set:** 166 questions with public annotations; **~300** held for leaderboard ([§1](https://arxiv.org/pdf/2311.12983)).
- **Answer types:** number, short string, or comma-separated list; **one** correct answer ([§3.2](https://arxiv.org/pdf/2311.12983)).
- **Scoring:** quasi **exact match** after normalization tied to answer type ([§3.2](https://arxiv.org/pdf/2311.12983)).
- **Prompt template:** `FINAL ANSWER: …` — see `eval/gaia/paper/system-prompt.txt` (Figure 2).

Public examples in **Figure 1** (used in-repo only to validate the scorer, not to claim model scores):

| Level | Illustrative answer shape |
|-------|---------------------------|
| 1 | Integer (`90`) |
| 2 | Signed decimal (`+4.6`) |
| 3 | `LastName; minutes` (`White; 5876`) |

Running those tasks against a **live agent** requires web/files; CI uses a **stub solver** to verify the pipeline only.

## Repository layout

```
eval/gaia/
  paper/system-prompt.txt
  fixtures/pilot/              # synthetic CI smoke
  fixtures/paper-examples/     # arXiv Fig. 1 (public)
  schema/task.schema.json
  results/                     # gitignored outputs
  data/                        # optional HF cache (gitignored)
scripts/eval/gaia/
  score.py                     # §3.2 quasi-exact-match + FINAL ANSWER extraction
  run-pilot.sh / run-paper-examples.sh
  load_hf.py                   # optional
deploy/eval/gaia.env.example
```

## Commands

```bash
make test-agent-eval-registry
make test-agent-eval-pilot
make test-agent-eval-paper
```

## Wiring agents (MCP)

1. Prefix with `eval/gaia/paper/system-prompt.txt`.
2. Run via **veil-engage** / **veil-graph** MCP (secure profile for network).
3. Collect `FINAL ANSWER:` lines → JSONL predictions.
4. `python3 scripts/eval/gaia/score.py --tasks … --predictions …`

Do not weaken engage hardening for eval.

## Controls

| Control | Meaning |
|---------|---------|
| `VEIL-EVAL-001` | Pilot + paper-example harness in CI |
| `VEIL-EVAL-002` | No gated HF data in git; arXiv is normative reference |

## Related

- [external-agent-store.md](external-agent-store.md)
- [engage-agentic-threats.md](engage-agentic-threats.md)
- [mcp-agents.md](../agents/mcp-agents.md)
