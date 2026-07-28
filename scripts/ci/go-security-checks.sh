#!/usr/bin/env bash
# Runs gosec, govulncheck and golangci-lint against every Go module in the
# repo (each go.mod scanned standalone, GOWORK=off — Go workspace mode does
# not let `./...` reach across module boundaries). One script, one loop,
# reused by a single CI job instead of a 63-way job matrix, and runnable
# locally the same way CI runs it:
#
#   scripts/ci/go-security-checks.sh
#
# Writes gosec SARIF per module to gosec-sarif-raw/<module-slug>.sarif, then
# merges them into a single gosec-sarif/gosec.sarif with one "run" (GitHub
# Actions stopped combining multiple SARIF runs uploaded under the same
# category as of 2025-07-21 — see
# https://github.blog/changelog/2025-07-21-code-scanning-will-stop-combining-multiple-sarif-runs-uploaded-in-the-same-sarif-file/
# — so upload-sarif must point at one run, not a directory of 21). Exits
# non-zero if any module has a real gosec finding, a govulncheck-reported
# vulnerability, or a golangci-lint issue.
set -u
cd "$(dirname "$0")/../.."
REPO_ROOT="$PWD"

mkdir -p gosec-sarif-raw gosec-sarif
fail=0

modules=$(find . -maxdepth 4 -name go.mod -not -path './var/*' | sed 's#/go.mod##; s#^\./##' | sort)

for mod in $modules; do
  slug=$(echo "$mod" | tr '/' '_')

  echo "::group::gosec: $mod"
  report="$REPO_ROOT/gosec-sarif-raw/${slug}.sarif"
  (cd "$mod" && GOWORK=off gosec -exclude-dir=.cache -exclude-dir=vendor -exclude-generated -fmt sarif -out "$report" ./...) >/dev/null 2>&1
  # gosec's own exit code is not trustworthy as a gate: on some packages its SSA
  # analyser panics with an internal error unrelated to any finding (seen on
  # packages using generics) and it still exits 1 with zero actual results.
  # Always parse the SARIF it produces and gate on real finding count instead.
  if [ -f "$report" ]; then
    count=$(python3 -c "import json; d=json.load(open('$report')); print(sum(len(r.get('results', [])) for r in d.get('runs', [])))")
    if [ "$count" -gt 0 ]; then
      python3 - "$report" "$mod" <<'PYEOF'
import json, sys
report, mod = sys.argv[1], sys.argv[2]
d = json.load(open(report))
for run in d.get("runs", []):
    for r in run.get("results", []):
        loc = r.get("locations", [{}])[0].get("physicalLocation", {}).get("artifactLocation", {}).get("uri", "?")
        line = r.get("locations", [{}])[0].get("physicalLocation", {}).get("region", {}).get("startLine", "?")
        msg = (r.get("message", {}).get("text") or "").splitlines()[0][:160]
        print(f"[{r.get('ruleId', '?')}] {mod}/{loc}:{line} — {msg}")
PYEOF
      echo "::error::gosec: ${count} finding(s) in $mod — see above"
      fail=1
    fi
    echo "[gosec] $mod findings=${count}"
  else
    echo "::error::gosec did not produce $report for $mod — refusing to silently pass a SAST gate with no real report"
    fail=1
  fi
  echo "::endgroup::"

  echo "::group::govulncheck: $mod"
  if ! (cd "$mod" && GOWORK=off govulncheck ./...); then
    echo "::error::govulncheck: vulnerability found in $mod — see above"
    fail=1
  fi
  echo "::endgroup::"

  echo "::group::golangci-lint: $mod"
  if ! (cd "$mod" && GOWORK=off golangci-lint run --config "$REPO_ROOT/.golangci.yml" ./...); then
    echo "::error::golangci-lint: issue(s) found in $mod — see above"
    fail=1
  fi
  echo "::endgroup::"
done

echo "::group::merge gosec SARIF"
python3 - "$REPO_ROOT/gosec-sarif-raw" "$REPO_ROOT/gosec-sarif/gosec.sarif" <<'PYEOF'
import glob, json, os, sys

raw_dir, out_path = sys.argv[1], sys.argv[2]
merged = {
    "version": "2.1.0",
    "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
    "runs": [{"tool": {"driver": {"name": "gosec", "rules": []}}, "results": []}],
}
run = merged["runs"][0]
seen_rule_ids = set()

for path in sorted(glob.glob(os.path.join(raw_dir, "*.sarif"))):
    mod = os.path.basename(path)[: -len(".sarif")].replace("_", "/")
    d = json.load(open(path))
    for src_run in d.get("runs", []):
        for rule in src_run.get("tool", {}).get("driver", {}).get("rules", []):
            if rule.get("id") not in seen_rule_ids:
                seen_rule_ids.add(rule.get("id"))
                run["tool"]["driver"]["rules"].append(rule)
        for result in src_run.get("results", []):
            for loc in result.get("locations", []):
                art = loc.get("physicalLocation", {}).get("artifactLocation")
                if art and "uri" in art and not art["uri"].startswith(mod + "/"):
                    art["uri"] = f"{mod}/{art['uri']}"
            run["results"].append(result)

json.dump(merged, open(out_path, "w"))
print(f"merged {len(run['results'])} gosec result(s) from {raw_dir} into {out_path}")
PYEOF
echo "::endgroup::"

exit $fail
