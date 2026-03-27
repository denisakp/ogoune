SHELL := /bin/sh

BINARY := dist/ogoune

.PHONY: build build-be build-fe test test-be test-fe lint clean docker

build: build-fe build-be

build-be:
	mkdir -p dist
	go build -o $(BINARY) ./cmd/api/main.go

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
