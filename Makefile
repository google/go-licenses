TMP = ./.tmp
RESULTS = $(TMP)/results
ASSETS = assets
DBASSET = $(ASSETS)/licenses.db
BIN = $(abspath $(TMP)/bin)
COVER_REPORT = $(RESULTS)/cover.report
COVER_TOTAL = $(RESULTS)/cover.total
LINTCMD = $(BIN)/golangci-lint run --tests=false --config .golangci.yaml

BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)
COVERAGE_THRESHOLD := 34

RELEASE_CMD=$(BIN)/goreleaser --rm-dist

ifndef TMP
    $(error TMP is not set)
endif

ifndef BIN
    $(error BIN is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef

.PHONY: all bootstrap lint lint-fix unit coverage help test clean ci-build-snapshot-packages ci-plugs-out-test ci-test-linux-run ci-test-linux-arm-run ci-test-mac-run ci-test-mac-arm-run ci-release

all: lint test ## Run all checks (linting, unit tests, and integration tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

test: unit ## Run all tests (currently only unit)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'

bootstrap: ## Download and install all project dependencies (+ prep tooling in the ./.tmp dir)
	$(call title,Downloading dependencies)
	@mkdir -p $(TMP) $(RESULTS) $(BIN) || exit 1
	go mod download || exit 1
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % env GOBIN=$(BIN) go install % || exit 1
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN) v1.50.1 || exit 1
	cd $(TMP) && curl -sLO https://github.com/markbates/pkger/releases/download/v0.17.0/pkger_0.17.0_$(shell uname)_x86_64.tar.gz && \
		tar -xzvf pkger_0.17.0_$(shell uname)_x86_64.tar.gz pkger && \
		mv pkger $(BIN) || exit 1
	GOBIN=$(BIN) go install github.com/goreleaser/goreleaser@v1.3.1 || exit 1

$(DBASSET):
	$(call title,Building assets)
	@mkdir -p $(ASSETS) || exit 1
	$(BIN)/license_serializer -output $(ASSETS) || exit 1

pkged.go: $(DBASSET)
	$(BIN)/pkger || exit 1

lint: ## Run gofmt + golangci-lint checks
	$(call title,Running linters)
	@FILES_WITH_ISSUES="$(shell gofmt -l -s .)"; \
	if [ -n "$$FILES_WITH_ISSUES" ]; then \
		echo "The following files have gofmt issues:"; \
		echo "$$FILES_WITH_ISSUES"; \
		exit 1; \
	fi
	@printf "Running golangci-lint...\n"
	$(LINTCMD) || exit 1

lint-fix: ## Auto-format all source code + run golangci-lint fixers
	$(call title,Running lint fixers)
	@echo "Running gofmt to auto-format code..."
	@gofmt -w -s . || exit 1
	@echo "Running golangci-lint with --fix..."
	$(LINTCMD) --fix || exit 1

unit: ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -coverprofile $(COVER_REPORT) ./...
	@go tool cover -func $(COVER_REPORT) | grep total | awk '{print substr($$3, 1, length($$3)-1)}' > $(COVER_TOTAL)
	@echo "Coverage: $$(cat $(COVER_TOTAL))"
	@if [ $$(echo "$$(cat $(COVER_TOTAL)) >= $(COVERAGE_THRESHOLD)" | bc -l) -ne 1 ]; then \
		echo "$(RED)$(BOLD)Failed coverage quality gate (> $(COVERAGE_THRESHOLD)%)$(RESET)" && false; \
	fi

ci-build-snapshot-packages: pkged.go
	$(RELEASE_CMD) --snapshot --skip-publish

ci-plugs-out-test:
	docker run \
		-v //var/run/docker.sock://var/run/docker.sock \
		-v ${PWD}://src \
		-w //src \
		golang:latest \
			/bin/bash -x -c "\
				./dist/go-licenses_linux_amd64/golicenses version && \
				./dist/go-licenses_linux_amd64/golicenses list github.com/khulnasoft/go-licenses && \
				./dist/go-licenses_linux_amd64/golicenses check github.com/khulnasoft/go-licenses \
			"

ci-test-linux-run:
	chmod 755 ./dist/go-licenses_linux_amd64/golicenses && \
	./dist/go-licenses_linux_amd64/golicenses version && \
	./dist/go-licenses_linux_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-test-linux-arm-run:
	chmod 755 ./dist/go-licenses_linux_arm64/golicenses && \
	./dist/go-licenses_linux_arm64/golicenses version && \
	./dist/go-licenses_linux_arm64/golicenses list github.com/khulnasoft/go-licenses

ci-test-mac-run:
	chmod 755 ./dist/go-licenses_darwin_amd64/golicenses && \
	./dist/go-licenses_darwin_amd64/golicenses version && \
	./dist/go-licenses_darwin_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-test-mac-arm-run:
	chmod 755 ./dist/go-licenses_darwin_arm64/golicenses && \
	./dist/go-licenses_darwin_arm64/golicenses version && \
	./dist/go-licenses_darwin_arm64/golicenses list github.com/khulnasoft/go-licenses

ci-release: pkged.go
	$(BIN)/goreleaser --rm-dist

clean: ## Clean build artifacts
	rm -rf dist .tmp $(RESULTS)
