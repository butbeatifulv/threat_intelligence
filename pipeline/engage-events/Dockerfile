FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY pkg/ pkg/
COPY pipeline/connector/go.mod pipeline/connector/go.sum pipeline/connector/
COPY pipeline/engage-events/go.mod pipeline/engage-events/go.sum pipeline/engage-events/
COPY pipeline/connector/ pipeline/connector/
COPY pipeline/engage-events/ pipeline/engage-events/
WORKDIR /src/pipeline/engage-events
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /engage-events-worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /engage-events-worker /engage-events-worker
ENTRYPOINT ["/engage-events-worker"]
