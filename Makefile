SHELL := /bin/sh

BINARY := dist/ogoune
SQLC_VERSION := v1.27.0

.PHONY: build build-be build-fe test test-be test-fe lint clean docker swag run-ci license-audit sqlc-bin sqlc-generate sqlc-check

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

build-fe:
	cd web && pnpm build

test: test-be test-fe

test-be:
	go test -race ./...

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
