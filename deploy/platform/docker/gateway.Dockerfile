# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY platform/gateway/ platform/gateway/
COPY pkg/ pkg/
ENV CGO_ENABLED=0
WORKDIR /build/platform/gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/veil-gateway ./cmd/veil-gateway

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/veil-gateway /veil-gateway
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/veil-gateway", "healthcheck"]
ENTRYPOINT ["/veil-gateway"]
