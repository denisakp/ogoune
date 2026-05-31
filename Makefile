SHELL := /bin/sh

BINARY := dist/ogoune
SQLC_VERSION := v1.27.0

.PHONY: build build-be build-fe test test-be test-be-pg test-be-bench test-fe lint clean docker swag run-ci license-audit sqlc-bin sqlc-generate sqlc-check migrations-drift-check

build: build-fe build-be

build-be: sqlc-check
	mkdir -p dist
	go build -o $(BINARY) ./cmd/api/main.go

sqlc-bin:
	@command -v sqlc >/dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

sqlc-generate: sqlc-bin
	sqlc generate -f sqlc.yaml

sqlc-check: sqlc-bin
	@sqlc generate -f sqlc.yaml
	@drift=$$(git status --porcelain -- internal/repository/sqlc/pg internal/repository/sqlc/sqlite \
		| grep -Ev '^A  ' || true); \
	if [ -n "$$drift" ]; then \
		echo "sqlc drift: run 'make sqlc-generate' and commit the result"; \
		printf '%s\n' "$$drift"; \
		exit 1; \
	fi

migrations-drift-check:
	go run ./cmd/migrations-drift-check

build-fe:
	cd web && pnpm build

test: test-be test-fe

test-be:
	go test -race ./...

# Run the backend tests with Postgres enabled. Provisioning is owned by
# testcontainers-go inside internal/repository/internaltest; the helper
# boots postgres:16-alpine on first use and tears it down at process exit.
# Skips gracefully when Docker is not reachable.
test-be-pg:
	@if ! docker info >/dev/null 2>&1; then \
		echo "Docker not available — skipping Postgres tests"; \
		exit 0; \
	fi
	go test -race -timeout 300s ./internal/repository/store/... ./internal/repository/internaltest/...

# Paired benches (spec 049): GORM vs sqlc p95 ratio gates on Resource.List
# and Incident.GetIncidentStats. Runs WITHOUT -race (race detector inflates
# p95 ~10× and drowns the signal). SQLite only — paired ratio is dialect-
# invariant; adding the Postgres testcontainer would double bench time
# without changing the signal.
#
# Output: each bench emits a structured `paired_bench …` line for CI capture.
# Gate: default warn-and-pass; set PAIRED_BENCH_STRICT=true to escalate
# ratio > 1.10 to a hard failure.
test-be-bench:
	go test -bench=Paired -benchtime=1x -run=^$$ -count=1 \
	  ./internal/repository/store/... | tee bench-output.txt

test-fe:
	cd web && pnpm test

lint:
	go vet ./...
	cd web && pnpm lint

clean:
	rm -rf dist
	rm -rf web/dist

docker:
	docker build -t ogoune:test .

swag:
	swag init -g cmd/api/main.go --output docs --parseDependency --parseInternal

run-ci:
	@echo "=== Lint ==="
	go vet ./...
	cd web && pnpm lint
	@echo "=== Backend tests (race + timeout, like CI) ==="
	go test -v -race -timeout 120s ./...
	@echo "=== Frontend tests ==="
	cd web && pnpm test
	@echo "=== Build ==="
	$(MAKE) build
	@echo "=== CI local: ALL PASSED ==="

license-audit:
	@echo "=== SPDX coverage guard ==="
	scripts/license/check-spdx.sh
	@echo "=== Runtime-deps license guard ==="
	scripts/license/check-deps.sh
	@echo "=== Docs AGPL-drift guard ==="
	scripts/license/check-docs.sh
	@echo "=== License audit: ALL PASSED ==="
