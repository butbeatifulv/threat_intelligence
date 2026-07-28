#!/usr/bin/env bash
# Cypher counts for NVD CWE/CPE relationships after ingest.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

cd "${VEIL_ROOT}"
compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
  "MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
   MATCH (v:Vulnerability)-[:AFFECTS]->(:CPE) RETURN count(*) AS affects;
   MATCH (n:CPE) RETURN count(n) AS cpe_nodes;"
