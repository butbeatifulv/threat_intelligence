# NVD helpers (pipeline only)

NVD CVE API 2.0 parsing for pipeline NED enrich. Harvest publishes raw `scrape_nvd_page` only; CWE/CPE extraction runs in [ned/internal/sources/vuln/enrich](../../ned/internal/sources/vuln/enrich/).

- `parse` — extract vulnerabilities from a JSON page
- `map` — map parsed rows into the canonical vulnerability shape
