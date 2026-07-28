# syntax=docker/dockerfile:1
# Playwright + Chromium browser sidecar for discovery browser crawl.
FROM mcr.microsoft.com/playwright:v1.49.1-jammy
WORKDIR /app
COPY discovery/cmd/browser-agent/package.json discovery/cmd/browser-agent/index.mjs ./
RUN npm install --omit=dev && npx playwright install chromium
ENV DISCOVERY_BROWSER_SIDECAR_LISTEN=:8910
EXPOSE 8910
HEALTHCHECK CMD curl -fsS http://127.0.0.1:8910/health || exit 1
CMD ["node", "index.mjs"]
