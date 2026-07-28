# terraform-provider-cvp Makefile
# All Go/Terraform commands run inside the Podman dev container.
# No host Go toolchain is required — see docs/development.md.

COMPOSE_FILE := deployments/compose/compose.dev.yml
DC   := podman-compose -f $(COMPOSE_FILE)
EXEC := $(DC) exec -T dev

# Provider identity (Terraform Registry namespace/type).
NAMESPACE  := ioplane
TYPE       := cvp
BINARY     := terraform-provider-$(TYPE)

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || cat VERSION 2>/dev/null || echo dev)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
  -X main.version=$(VERSION) \
  -X main.commit=$(GIT_COMMIT) \
  -X main.date=$(BUILD_DATE)

REPORT_DIR := reports

.PHONY: help all verify up down restart logs shell \
        build install \
        test test-v test-run testacc \
        lint lint-fix lint-md lint-yaml lint-spell lint-docs tffmt tffmt-check \
        vulncheck osv-scan gosec \
        proto proto-lint proto-breaking \
        docs docs-check \
        tidy download \
        clean versions

help:
	@echo "terraform-provider-cvp Makefile targets"
	@echo "======================================="
	@echo "Lifecycle:   up down restart logs shell"
	@echo "Build:       build install"
	@echo "Test:        test test-v testacc"
	@echo "Quality:     lint lint-fix vulncheck osv-scan gosec"
	@echo "Proto:       proto proto-lint proto-breaking"
	@echo "Docs:        docs docs-check lint-md lint-yaml lint-spell lint-docs"
	@echo "Terraform:   tffmt tffmt-check"
	@echo "Deps:        tidy download"
	@echo "Gates:       all (build+test+lint+docs)  verify (all + testacc)"
	@echo "Misc:        clean versions"

# === Lifecycle ===

up:
	$(DC) up -d --build

down:
	$(DC) down

restart: down up

logs:
	$(DC) logs -f dev

shell:
	$(DC) exec dev bash

# === Build ===

build:
	$(EXEC) go build -ldflags='$(LDFLAGS)' -o bin/$(BINARY) .

# Install into the local Terraform plugin dir for manual `dev_overrides` testing.
install:
	$(EXEC) go install -ldflags='$(LDFLAGS)' .

# === Test ===

# Unit tests (no live CVP). Race detector always on.
test:
	$(EXEC) go test ./... -race -count=1

test-v:
	$(EXEC) go test ./... -race -count=1 -v

test-run:
	@test -n "$(RUN)" || (echo "Usage: make test-run RUN=TestX PKG=./internal/x"; exit 1)
	$(EXEC) go test -run '$(RUN)' $(PKG) -race -count=1 -v

# Acceptance tests (TF_ACC=1). Requires live CVP endpoint + credentials in the
# environment (CVP_ENDPOINT / CVP_AUTH_METHOD / CVP_TOKEN — sourced from gopass,
# see AGENTS.md). Never run in the default PR pipeline.
testacc:
	$(EXEC) env TF_ACC=1 go test ./... -race -count=1 -timeout 120m -v

# === Quality ===

lint:
	$(EXEC) golangci-lint run ./...

lint-fix:
	$(EXEC) golangci-lint run --fix ./...

vulncheck:
	$(EXEC) govulncheck ./...

osv-scan:
	$(EXEC) osv-scanner scan --recursive .

gosec:
	$(EXEC) golangci-lint run --enable-only gosec ./...

# === Protobuf (custom CVP client, ADR 0005) ===

# The targets are no-ops until protos are vendored (P1), so they stay green in
# the meantime instead of failing on an empty module.
HAVE_PROTO := $(shell find api/proto -name '*.proto' 2>/dev/null | head -1)

# Generate Go gRPC stubs from the vendored protos under api/proto → internal/pb.
proto:
	@if [ -n "$(HAVE_PROTO)" ]; then $(EXEC) buf generate; else echo "proto: no .proto files yet (P1) — skipping"; fi

proto-lint:
	@if [ -n "$(HAVE_PROTO)" ]; then $(EXEC) buf lint; else echo "proto-lint: no .proto files yet (P1) — skipping"; fi

# Compare the working protos against the last committed revision.
proto-breaking:
	@if [ -n "$(HAVE_PROTO)" ]; then $(EXEC) buf breaking --against '.git#branch=main'; else echo "proto-breaking: no .proto files yet (P1) — skipping"; fi

# === Terraform formatting (examples) ===

tffmt:
	$(EXEC) terraform fmt -recursive examples

tffmt-check:
	$(EXEC) terraform fmt -check -recursive examples

# === Docs ===

# Generate registry documentation from schema + templates/examples.
docs:
	$(EXEC) tfplugindocs generate --provider-name $(TYPE)

docs-check:
	$(EXEC) tfplugindocs validate --provider-name $(TYPE)

lint-md:
	$(EXEC) markdownlint-cli2 "**/*.md" "#node_modules" "#bin" "#dist"

lint-yaml:
	$(EXEC) yamllint -c .yamllint.yaml .

lint-spell:
	$(EXEC) cspell --no-progress --no-summary --config .cspell.json "**/*.md" "**/*.go"

lint-docs: lint-md lint-yaml lint-spell docs-check

# === Deps ===

tidy:
	$(EXEC) go mod tidy

download:
	$(EXEC) go mod download

# === Aggregate gates ===

# Pre-PR gate: everything that does NOT need a live CVP.
all: build test lint tffmt-check lint-docs vulncheck

# Full gate: `all` plus live acceptance tests against the netlab2 CVP lab.
verify: all testacc
	@echo "verify: all gates green"

# === Clean ===

clean:
	rm -rf bin/ dist/ $(REPORT_DIR) coverage.out coverage.html

# === Info ===

versions:
	@echo "=== Go ==="            && $(EXEC) go version
	@echo "=== Terraform ==="     && $(EXEC) terraform version
	@echo "=== OpenTofu ==="      && $(EXEC) tofu version
	@echo "=== Terragrunt ==="    && $(EXEC) terragrunt --version
	@echo "=== golangci-lint ===" && $(EXEC) golangci-lint version --short
	@echo "=== govulncheck ==="   && $(EXEC) govulncheck -version 2>/dev/null || echo installed
	@echo "=== tfplugindocs ==="  && $(EXEC) tfplugindocs --version 2>/dev/null || echo installed
