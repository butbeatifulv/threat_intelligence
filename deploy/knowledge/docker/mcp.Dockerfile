# syntax=docker/dockerfile:1
# MCP stdio + optional Streamable HTTP (MCP_HTTP_ENABLED=1).
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY knowledge/ knowledge/
COPY pkg/ pkg/
ENV GOWORK=/build/knowledge/go.work
ENV CGO_ENABLED=0
WORKDIR /build/knowledge/serve
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/mcp ./cmd/mcp

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/mcp /mcp
USER nonroot:nonroot
EXPOSE 8091
HEALTHCHECK --interval=20s --timeout=5s --start-period=30s --retries=3 \
  CMD ["/mcp", "healthcheck"]
ENTRYPOINT ["/mcp"]
