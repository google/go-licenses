TMP = ./.tmp
RESULTS = $(TMP)/results
ASSETS = assets
DBASSET = $(ASSETS)/licenses.db
BIN = $(abspath $(TMP)/bin)
COVER_REPORT = $(RESULTS)/cover.report
COVER_TOTAL = $(RESULTS)/cover.total
LINTCMD = $(BIN)/golangci-lint run --tests=false --config .golangci.yaml

# Color definitions with checks for OS
ifeq ($(OS),Windows_NT)
    BOLD := ""
    PURPLE := ""
    GREEN := ""
    CYAN := ""
    RED := ""
    RESET := ""
else
    BOLD := $(shell tput -T linux bold)
    PURPLE := $(shell tput -T linux setaf 5)
    GREEN := $(shell tput -T linux setaf 2)
    CYAN := $(shell tput -T linux setaf 6)
    RED := $(shell tput -T linux setaf 1)
    RESET := $(shell tput -T linux sgr0)
endif

TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)
COVERAGE_THRESHOLD := 55

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

.PHONY: all bootstrap lint lint-fix unit coverage help test clean

all: lint test ## Run all checks (linting, unit tests, and integration tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

test: unit ## Run all tests (currently only unit)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'

bootstrap: ## Download and install all project dependencies (+ prep tooling in the ./.tmp dir)
	$(call title,Downloading dependencies)
	mkdir -p $(TMP) $(RESULTS) $(BIN)  # Create necessary directories
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % env GOBIN=$(BIN) go install %
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN) v1.47.2
	cd $(TMP) && curl -sLO https://github.com/markbates/pkger/releases/download/v0.17.0/pkger_0.17.0_$(shell uname)_x86_64.tar.gz && \
		tar -xzvf pkger_0.17.0_$(shell uname)_x86_64.tar.gz pkger && \
		mv pkger $(BIN)
	GOBIN=$(BIN) go install github.com/goreleaser/goreleaser@v1.3.1

$(DBASSET):
	$(call title,Building assets)
	mkdir -p $(ASSETS)
	$(BIN)/license_serializer -output $(ASSETS)

pkged.go: $(DBASSET)
	$(BIN)/pkger

lint: ## Run gofmt + golangci lint checks
	$(call title,Running linters)
	@printf "files with gofmt issues: [$(shell gofmt -l -s .)]\n"
	@test -z "$(shell gofmt -l -s .)"
	$(LINTCMD)

lint-fix: ## Auto-format all source code + run golangci lint fixers
	$(call title,Running lint fixers)
	gofmt -w -s .
	$(LINTCMD) --fix

unit: ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -coverprofile $(COVER_REPORT) ./...
	@go tool cover -func $(COVER_REPORT) | grep total | awk '{print substr($$3, 1, length($$3)-1)}' > $(COVER_TOTAL)
	@echo "Coverage: $$(cat $(COVER_TOTAL))"
	@if [ $$(echo "$$(cat $(COVER_TOTAL)) >= $(COVERAGE_THRESHOLD)" | bc -l) -ne 1 ]; then \
		echo "$(RED)$(BOLD)Failed coverage quality gate (> $(COVERAGE_THRESHOLD)%)$(RESET)" && false; \
	fi

ci-build-snapshot-packages: pkged.go
	$(RELEASE_CMD) \
		--snapshot \
		--skip-publish 

ci-plugs-out-test:
	docker run \
		-v //var/run/docker.sock://var/run/docker.sock \
		-v ${PWD}://src \
		-w //src \
		golang:latest \
			/bin/bash -x -c "\
				./dist/linux-build_linux_amd64/golicenses version && \
				./dist/linux-build_linux_amd64/golicenses list github.com/khulnasoft/go-licenses && \
				./dist/linux-build_linux_amd64/golicenses check github.com/khulnasoft/go-licenses \
			"

ci-test-linux-run:
	chmod 755 ./dist/linux-build_linux_amd64/golicenses && \
	./dist/linux-build_linux_amd64/golicenses version && \
	./dist/linux-build_linux_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-test-linux-arm-run:
	chmod 755 ./dist/linux-build_linux_arm64/golicenses && \
	./dist/linux-build_linux_arm64/golicenses version && \
	./dist/linux-build_linux_arm64/golicenses list github.com/khulnasoft/go-licenses

ci-test-mac-run:
	chmod 755 ./dist/darwin-build_darwin_amd64/golicenses && \
	./dist/darwin-build_darwin_amd64/golicenses version && \
	./dist/darwin-build_darwin_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-test-mac-arm-run:
	chmod 755 ./dist/darwin-build_darwin_arm64/golicenses && \
	./dist/darwin-build_darwin_arm64/golicenses version && \
	./dist/darwin-build_darwin_arm64/golicenses list github.com/khulnasoft/go-licenses

ci-test-windows-run:
	@echo "Running Windows tests..."
	@powershell -Command "Start-Process -NoNewWindow -File ./dist/windows-build_windows_amd64/golicenses.exe -ArgumentList 'version' -Wait"
	@powershell -Command "Start-Process -NoNewWindow -File ./dist/windows-build_windows_amd64/golicenses.exe -ArgumentList 'list github.com/khulnasoft/go-licenses' -Wait"

ci-release: pkged.go
	$(BIN)/goreleaser --rm-dist

clean:
	rm -rf dist
	rm -rf .tmp
