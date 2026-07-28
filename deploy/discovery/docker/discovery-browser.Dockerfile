# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY . .
RUN cd discovery/browser && go build -o /discovery-browser ./cmd/serve

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
COPY --from=build /discovery-browser /usr/local/bin/discovery-browser
ENV DISCOVERY_BROWSER_LISTEN=:8920
ENV DISCOVERY_BROWSER_SIDECAR_URL=http://discovery-browser-agent:8910
EXPOSE 8920
HEALTHCHECK CMD curl -fsS http://127.0.0.1:8920/health || exit 1
CMD ["discovery-browser"]
