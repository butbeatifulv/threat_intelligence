# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY knowledge/ knowledge/
COPY pkg/ pkg/
ENV GOWORK=/build/knowledge/go.work
WORKDIR /build/knowledge/ingest
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/ingest_worker ./cmd/ingest_worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/ingest_worker /usr/local/bin/ingest_worker
USER nobody
ENTRYPOINT ["/usr/local/bin/ingest_worker"]
