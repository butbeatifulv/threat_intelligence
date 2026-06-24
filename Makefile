.PHONY: test-discovery test-discovery-p7c skills-index check-skills-index corpus-import check-corpus-mappings procedures-index check-procedures-index test-pipeline test-pipeline-p7d test-graph test-graph-ingest-p7e test-graph-serve-p7f test-graph-serve test-graph-read-smoke test-graph-engage-category test-engage-events-pipeline test-platform-p0 test-platform-p7 test-platform-closed-loop test-platform-full-loop test-platform-p3 test-platform-p4 test-platform-mcp-gateway test-platform-unified-edge graph-pack-export graph-pack-build graph-pack-publish test-smoke check-graph-version bump-graph-patch agents-list agents-render deploy-helm-template deploy-ansible-check sync-github-metadata external-clone-agent-store test-agent-eval-registry test-agent-eval-pilot test-agent-eval-paper test-pkg-shared test-pkg-domain test-pkg-all test-pkg-cover test-pkg-cover-strict test-knowledge test-knowledge-serve pentest-veil-mcp

# Shared pkg contracts (harvest, commit, natsjet, auth, engage/events)
test-pkg-shared:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/... ./natsjet/...
	cd pkg/api && env -u GOWORK go test ./...
	cd pkg/auth && env -u GOWORK go test ./...
	cd pkg/mcp && env -u GOWORK go test ./...
	cd pkg/engage && env -u GOWORK go test ./...
	cd pkg/exec && env -u GOWORK go test ./...

# pkg domain contour (meta-layer + per-source + engage domain + auth httpmiddleware)
test-pkg-domain:
	cd pkg && env -u GOWORK go test ./domain/... ./ti/... ./vuln/domain/... ./lola/domain/... \
		./ds/domain/... ./sbom/domain/... ./nuclei/domain/... ./coderules/domain/... ./decision/... \
		./playbook/...
	cd pkg/engage && env -u GOWORK go test ./domain/... ./contract/... ./toolid/...
	cd pkg/auth && env -u GOWORK go test ./httpmiddleware/...

# Full pkg unit tests (root module + engage submodule + api/auth/mcp/exec)
test-pkg-all:
	cd pkg && env -u GOWORK go test ./...
	cd pkg/engage && env -u GOWORK go test ./...
	cd pkg/api && env -u GOWORK go test ./...
	cd pkg/auth && env -u GOWORK go test ./...
	cd pkg/mcp && env -u GOWORK go test ./...
	cd pkg/exec && env -u GOWORK go test ./...

# T0/T2 gates: presence + coverage floors (see docs/development/pkg-test-coverage.md)
test-pkg-cover:
	chmod +x ./scripts/test/pkg-cover.sh
	./scripts/test/pkg-cover.sh

# T3 gate: 100% statement coverage on logic packages
test-pkg-cover-strict:
	chmod +x ./scripts/test/pkg-cover.sh ./scripts/test/pkg-cover-strict.sh
	./scripts/test/pkg-cover-strict.sh

# P7 gate: pkg + bus + layer unit tests (wave-1 parallel branches merged)
test-platform-p7: test-pkg-cover-strict test-platform-p0 test-discovery-p7c test-pipeline-p7d test-graph-ingest-p7e test-graph-serve-p7f

# GOWORK may point at discovery/go.work in the shell; each target uses the matching workspace.
test-platform-p0: test-pkg-shared
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./connector/nats/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./ned/internal/consumer/... ./ned/internal/dedup/...
	cd knowledge && env GOWORK=$$(pwd)/go.work go test ./ingest/internal/ingest/...

test-platform-closed-loop:
	chmod +x ./scripts/test/smoke-platform-closed-loop.sh
	./scripts/test/smoke-platform-closed-loop.sh

test-platform-full-loop:
	chmod +x ./scripts/test/smoke-platform-full-loop.sh
	./scripts/test/smoke-platform-full-loop.sh

test-platform-p3: test-platform-p0 test-platform-closed-loop

test-platform-p4: test-platform-p0 test-platform-full-loop

# P12f: unified MCP HTTP aggregator (graph + engage backends)
test-platform-mcp-gateway:
	cd platform/mcp-gateway && env -u GOWORK go test ./...

# P12i: unified TLS nginx edge (Docker; skip with SMOKE_SKIP_UNIFIED_EDGE=1)
test-platform-unified-edge:
	chmod +x ./scripts/test/smoke-unified-edge.sh
	./scripts/test/smoke-unified-edge.sh

agents-list:
	chmod +x ./scripts/agents/list-manifest.sh
	./scripts/agents/list-manifest.sh

agents-render:
	chmod +x ./scripts/agents/render-task-prompt.sh
	@test -n "$(AGENT)" || (echo "usage: make agents-render AGENT=platform-implementer [PHASE=platform-p4b]" >&2; exit 1)
	./scripts/agents/render-task-prompt.sh "$(AGENT)" $(if $(PHASE),--phase $(PHASE),)

deploy-helm-template:
	@if command -v helm >/dev/null 2>&1; then \
		helm template veil deploy/helm/veil -f deploy/helm/veil/values.yaml \
			-f deploy/helm/veil/values-stage.yaml \
			--set global.imageTag=$${APP_VERSION:-v0.4.5}; \
	else echo "SKIP: helm not installed"; fi

deploy-ansible-check:
	@if command -v ansible-playbook >/dev/null 2>&1; then \
		ansible-playbook deploy/ansible/playbooks/site.yml -i deploy/ansible/inventories/stage --syntax-check; \
	else echo "SKIP: ansible-playbook not installed"; fi

sync-github-metadata:
	chmod +x ./scripts/housekeeping/sync-github-metadata.sh
	./scripts/housekeeping/sync-github-metadata.sh

external-clone-agent-store:
	chmod +x ./scripts/external/clone-agent-store.sh
	./scripts/external/clone-agent-store.sh

test-agent-eval-registry:
	python3 ./scripts/eval/agent-eval-registry-audit.py

test-agent-eval-pilot:
	chmod +x ./scripts/eval/gaia/run-pilot.sh ./scripts/eval/gaia/solvers/stub.sh
	./scripts/eval/gaia/run-pilot.sh

test-agent-eval-paper:
	chmod +x ./scripts/eval/gaia/run-paper-examples.sh ./scripts/eval/gaia/solvers/stub.sh
	./scripts/eval/gaia/run-paper-examples.sh

pentest-veil-mcp:
	chmod +x ./scripts/eval/pentest-veil-mcp.sh
	./scripts/eval/pentest-veil-mcp.sh

test-discovery:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/...
	cd discovery/pkg && env -u GOWORK go test ./...
	cd discovery && env GOWORK=$$(pwd)/go.work go test ./connector/... ./harvest/... ./browser/...
	cd discovery/harvest && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/scrape_worker
	cd discovery/browser && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/serve
	cd discovery/cmd/browser-agent && env GOWORK=$$(dirname $$(dirname $$(pwd)))/go.work go build -o /dev/null .

# P7c slice: TI feeds/helpers + shared discovery feeds (see veil_platform_p7 plan)
test-discovery-p7c:
	cd discovery && env GOWORK=$$(pwd)/go.work go test ./harvest/internal/sources/ti/... ./harvest/internal/sources/lola/... ./harvest/internal/sources/ds/... ./harvest/internal/feeds/...
	cd discovery/pkg && env -u GOWORK go test ./proxypool/...

test-pipeline-p7d:
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./pkg/nvd/map/... ./ned/internal/sources/appsec/... ./ned/internal/sources/ds/... ./ned/internal/sources/lola/...

test-graph-ingest-p7e:
	cd knowledge && env GOWORK=$$(pwd)/go.work go test ./ingest/internal/ingest/... ./ingest/internal/sources/ti/... ./ingest/internal/sources/vuln/... ./ingest/internal/sources/ds/... ./ingest/internal/sources/lola/... ./ingest/internal/appsec/...

test-graph-serve-p7f:
	cd knowledge && env GOWORK=$$(pwd)/go.work go test ./serve/internal/usecase/... ./connector/query/...

# Pentest execution targets → https://github.com/butbeautifulv/veneno (make test-engage)

test-engage-events-pipeline:
	chmod +x ./scripts/test/smoke-engage-events-pipeline.sh
	./scripts/test/smoke-engage-events-pipeline.sh

test-pipeline:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/... ./ti/...
	cd pipeline/pkg && env -u GOWORK go test ./nvd/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go build -o /dev/null ./connector/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./connector/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./ned/...
	cd pipeline/ned && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/pipeline_worker
	cd pipeline/engage-events && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/worker

test-knowledge:
	cd pkg && env -u GOWORK go test ./commit/... ./ti/...
	cd knowledge && env GOWORK=$$(pwd)/go.work go build -o /dev/null ./connector/...
	cd knowledge/ingest && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/ingest_worker
	cd knowledge/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./...
	cd knowledge/serve && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/api ./cmd/mcp

test-knowledge-serve:
	cd knowledge/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./... -race -count=1

# Deprecated aliases (remove after one release)
test-graph: test-knowledge

test-graph-serve: test-knowledge-serve

test-graph-read-smoke:
	./scripts/test/smoke-graph-read.sh

test-graph-engage-category:
	chmod +x ./scripts/test/smoke-graph-engage-category.sh
	./scripts/test/smoke-graph-engage-category.sh

graph-pack-export:
	./scripts/graph-pack/export-cypher.sh

graph-pack-build:
	./scripts/graph-pack/build.sh $(GRAPH_PACK_VERSION)

graph-pack-publish:
	./scripts/release/publish-graph-pack.sh

test-smoke:
	chmod +x ./scripts/test/smoke-discovery-e2e.sh
	./scripts/test/smoke-discovery-e2e.sh

check-graph-version:
	./scripts/release/check-graph-version-bump.sh

bump-graph-patch:
	./scripts/release/bump-graph-version.sh patch

corpus-import:
	chmod +x ./scripts/knowledge/corpus-import.sh
	./scripts/knowledge/corpus-import.sh

check-corpus-mappings:
	chmod +x ./scripts/knowledge/check-corpus-mappings.sh
	./scripts/knowledge/check-corpus-mappings.sh

skills-index:
	python3 ./scripts/knowledge/generate-cyber-skills-index.py

check-skills-index:
	python3 ./scripts/knowledge/generate-cyber-skills-index.py --check

procedures-index:
	python3 ./scripts/knowledge/extract-procedures-index.py

check-procedures-index:
	python3 ./scripts/knowledge/extract-procedures-index.py --check
