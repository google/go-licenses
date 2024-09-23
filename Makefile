TMP = ./.tmp
RESULTS = $(TMP)/results
ASSETS = assets
DBASSET = $(ASSETS)/licenses.db
DIST = ./dist
# note: go tools requires an absolute path
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
# the quality gate lower threshold for unit test total % coverage (by function statements)
COVERAGE_THRESHOLD := 55

ifndef TMP
    $(error TMP is not set)
endif

ifndef BIN
    $(error BIN is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef

.PHONY: all bootstrap lint lint-fix unit coverage help test

all: lint test ## Run all checks (linting, unit tests, and integration tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

test: unit ## Run all tests (currently only unit)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'

bootstrap: ## Download and install all project dependencies (+ prep tooling in the ./.tmp dir)
	$(call title,Downloading dependencies)
	# prep temp dirs
	mkdir -p $(TMP)
	mkdir -p $(RESULTS)
	mkdir -p $(BIN)
	# download install project dependencies + tooling
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % env GOBIN=$(BIN) go install %
	# install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN) v1.26.0
	# install pkger
	cd $(TMP) && curl -sLo pkger.tar.gz https://github.com/markbates/pkger/releases/download/v0.17.0/pkger_0.17.0_$(shell uname)_$(shell uname -m).tar.gz && \
		tar -xzvf pkger.tar.gz pkger && \
		mv pkger $(BIN)
	# install goreleaser
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh -s -- -b $(BIN) v0.138.0

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
	@go tool cover -func $(COVER_REPORT) | grep total |  awk '{print substr($$3, 1, length($$3)-1)}' > $(COVER_TOTAL)
	@echo "Coverage: $$(cat $(COVER_TOTAL))"
	@if [ $$(echo "$$(cat $(COVER_TOTAL)) >= $(COVERAGE_THRESHOLD)" | bc -l) -ne 1 ]; then echo "$(RED)$(BOLD)Failed coverage quality gate (> $(COVERAGE_THRESHOLD)%)$(RESET)" && false; fi

snapshot: pkged.go
	$(BIN)/goreleaser \
		--snapshot \
		--skip-publish \
		--rm-dist

# The following targets are all CI related

# note: since google's licenseclassifier requires the go tooling ('go list' from x/tools/go/packages) we need to use a golang image
ci-plugs-out-test:
	docker run \
		-v //var/run/docker.sock://var/run/docker.sock \
		-v /${PWD}://src \
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

ci-test-mac-run:
	chmod 755 ./dist/go-licenses_darwin_amd64/golicenses && \
	./dist/go-licenses_darwin_amd64/golicenses version && \
	./dist/go-licenses_darwin_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-test-windows-run:
	./dist/go-licenses_windows_amd64/golicenses version && \
	./dist/go-licenses_windows_amd64/golicenses list github.com/khulnasoft/go-licenses

ci-release: pkged.go
	$(BIN)/goreleaser --rm-dist

# all clean-related targets

.PHONY: clean
clean: clean-dist ## Remove anything with state

.PHONY: clean-dist
clean-dist:
	rm -rf $(DIST)