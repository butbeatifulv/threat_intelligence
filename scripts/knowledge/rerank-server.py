#!/usr/bin/env python3
"""Optional cross-encoder rerank sidecar for VEIL hybrid search."""
from __future__ import annotations

import argparse
from typing import Any

try:
    from fastapi import FastAPI
    from pydantic import BaseModel
    import uvicorn
except ImportError:
    FastAPI = None  # type: ignore


class RerankRequest(BaseModel):
    query: str
    docs: list[str]
    top_k: int = 10


def create_app() -> Any:
    if FastAPI is None:
        raise RuntimeError("install fastapi uvicorn sentence-transformers")
    app = FastAPI()
    model = None

    @app.on_event("startup")
    def load_model() -> None:
        nonlocal model
        from sentence_transformers import CrossEncoder

        model = CrossEncoder("BAAI/bge-reranker-v2-m3")

    @app.post("/rerank")
    def rerank(req: RerankRequest) -> dict[str, Any]:
        if model is None:
            return {"indices": list(range(min(req.top_k, len(req.docs))))}
        pairs = [[req.query, d] for d in req.docs]
        scores = model.predict(pairs)
        ranked = sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)
        return {"indices": ranked[: req.top_k]}

    return app


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--port", type=int, default=8092)
    args = ap.parse_args()
    uvicorn.run(create_app(), host=args.host, port=args.port)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
