# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY pkg/ pkg/
COPY pipeline/ pipeline/
ENV GOWORK=/build/pipeline/go.work
WORKDIR /build/pipeline/ned
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/pipeline_worker ./cmd/pipeline_worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/pipeline_worker /usr/local/bin/pipeline_worker
USER nobody
ENTRYPOINT ["/usr/local/bin/pipeline_worker"]
