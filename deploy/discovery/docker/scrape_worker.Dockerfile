# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY pkg/ pkg/
COPY discovery/ scrape/
ENV GOWORK=/build/discovery/go.work
WORKDIR /build/discovery/harvest
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/scrape_worker ./cmd/scrape_worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/scrape_worker /usr/local/bin/scrape_worker
COPY --from=build /build/discovery/harvest/internal/sources/ti/example.jsonl /app/example.jsonl
COPY --from=build /build/discovery/harvest/internal/sources/sbom/fixtures/cve_list_seed.txt /fixtures/cve_list_seed.txt
USER nobody
ENTRYPOINT ["/usr/local/bin/scrape_worker"]
