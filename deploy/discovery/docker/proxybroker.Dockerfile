# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY discovery/ .
ENV GOWORK=/src/go.work
WORKDIR /src/proxybroker
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/proxybroker ./cmd

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/proxybroker /usr/local/bin/proxybroker
USER nobody
ENTRYPOINT ["/usr/local/bin/proxybroker"]
