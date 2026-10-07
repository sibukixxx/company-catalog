.DEFAULT_GOAL := help

GOPATH ?= $(shell go env GOPATH)
TOOLS := $(GOPATH)/bin

# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION=v2.11.4

# E2E test 用 env (toscana スタイル: 明示的に enabled にしないと走らない)
HOUJIN_E2E_TEST_ENABLED ?= false
TEST_RUN ?=
E2E_RUN ?=

###############################################################################
### help
###############################################################################

.PHONY: help
# https://postd.cc/auto-documented-makefile/
help: ## Show this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###############################################################################
### setup
###############################################################################

.PHONY: setup
setup: install-tools ## Set up local development environment

.PHONY: install-tools
install-tools: install-golangci-lint install-gofumpt ## Install developer tools

.PHONY: install-golangci-lint
install-golangci-lint:
	@if [ ! -f $(TOOLS)/golangci-lint ] || ! $(TOOLS)/golangci-lint --version | grep -q -F "version $(shell echo $(GOLANGCI_LINT_VERSION) | sed 's/v//')"; then \
		curl -vvsSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS) $(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: install-gofumpt
install-gofumpt:
	@if [ ! -f $(TOOLS)/gofumpt ]; then \
		go install mvdan.cc/gofumpt@latest; \
	fi

###############################################################################
### コード整形
###############################################################################

.PHONY: fmt
fmt: install-gofumpt ## Format code (gofmt + gofumpt)
	gofmt -w .
	$(TOOLS)/gofumpt -extra -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: install-golangci-lint ## Run golangci-lint
	$(TOOLS)/golangci-lint run -c .golangci.yml

.PHONY: lint-fix
lint-fix: install-golangci-lint ## Run golangci-lint --fix
	$(TOOLS)/golangci-lint run -c .golangci.yml --fix

.PHONY: fix
fix: fmt lint-fix vet ## Apply all auto fixes

.PHONY: release-check
release-check: test vet build lint ## Run the deterministic OSS release gate

###############################################################################
### test
###############################################################################

.PHONY: test
test: ## Run unit tests (TEST_RUN=TestName to filter)
	go test -race -vet=all $(if $(TEST_RUN),-run $(TEST_RUN)) ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: test-e2e
test-e2e: ## Run E2E tests against the real 法人番号 Web-API (requires HOUJIN_APP_ID)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set. Export it before running test-e2e."; \
		exit 1; \
	fi
	HOUJIN_E2E_TEST_ENABLED=true \
		go test -tags=e2e -race -vet=all -count=1 $(if $(E2E_RUN),-run $(E2E_RUN)) ./test/e2e/...

.PHONY: test-all
test-all: test test-e2e ## Run unit tests and E2E tests

###############################################################################
### debug / E2E helpers
###############################################################################
#
# AppID を漏らさずにローカルデバッグするためのターゲット群。
# 値そのものは echo せず、長さ・有無のみを報告する。
# 想定フロー: check-env → ping → test-e2e-phase4 → test-e2e
#

PING_NUMBER ?= 1180301018771
SEARCH_NAME ?= トヨタ自動車
SEARCH_PREFECTURE ?= 23
DIFF_DAYS ?= 7
DIFF_PREFECTURE ?=

.PHONY: check-env
check-env: ## Show whether HOUJIN_APP_ID / HOUJIN_E2E_TEST_ENABLED are set (value is NOT printed)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "HOUJIN_APP_ID:             NOT SET"; \
	else \
		echo "HOUJIN_APP_ID:             SET ($$(printf '%s' "$$HOUJIN_APP_ID" | wc -c | tr -d ' ') chars)"; \
	fi
	@if [ "$$HOUJIN_E2E_TEST_ENABLED" = "true" ]; then \
		echo "HOUJIN_E2E_TEST_ENABLED:   true"; \
	else \
		echo "HOUJIN_E2E_TEST_ENABLED:   not enabled (current: '$$HOUJIN_E2E_TEST_ENABLED')"; \
	fi
	@if [ "$$HOUJIN_E2E_VERBOSE" = "true" ]; then \
		echo "HOUJIN_E2E_VERBOSE:        true"; \
	fi

.PHONY: ping
ping: ## Quick AppID liveness check via Lookup (defaults to トヨタ; override with PING_NUMBER=...)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set. Export it before running ping."; \
		exit 1; \
	fi
	@echo "Looking up corporate number: $(PING_NUMBER)"
	@echo "Do not log Config, unwrapped transport causes or raw request URLs (they may contain AppID)."
	go run ./examples/lookup_by_number $(PING_NUMBER)

.PHONY: search
search: ## Quick SearchByName check (SEARCH_NAME='...' SEARCH_PREFECTURE='23' で上書き)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set. Export it before running search."; \
		exit 1; \
	fi
	@echo "SearchByName: name=$(SEARCH_NAME) prefecture=$(SEARCH_PREFECTURE)"
	go run ./examples/search_by_name "$(SEARCH_NAME)" "$(SEARCH_PREFECTURE)"

.PHONY: diff
diff: ## Quick FetchUpdatesByDateRange check (DIFF_DAYS=7 DIFF_PREFECTURE='' で上書き; pagination 発生検出も表示)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set. Export it before running diff."; \
		exit 1; \
	fi
	go run ./examples/fetch_diff $(DIFF_DAYS) "$(DIFF_PREFECTURE)"

.PHONY: verify
verify: ## 取得検証フロー: Lookup + SearchByName + FetchUpdatesByDateRange を順に実行
	@$(MAKE) ping
	@echo "---"
	@$(MAKE) search
	@echo "---"
	@$(MAKE) diff

.PHONY: test-e2e-phase4
test-e2e-phase4: ## Run only Phase 4 E2E tests (pagination + cache)
	@$(MAKE) test-e2e E2E_RUN='TestE2EFetchUpdatesByDateRangeAggregatesDividedPages|TestE2EMemoryCacheHitAvoidsSecondAPICall'

.PHONY: test-e2e-verbose
test-e2e-verbose: ## Run E2E with HOUJIN_E2E_VERBOSE=true (extra t.Log output, e.g. cache hit details)
	HOUJIN_E2E_VERBOSE=true $(MAKE) test-e2e

.PHONY: debug-e2e
debug-e2e: ## One-shot debug flow: check env → verify (lookup+search+diff) → Phase 4 E2E
	@$(MAKE) check-env
	@echo "---"
	@$(MAKE) verify
	@echo "---"
	@$(MAKE) test-e2e-phase4

###############################################################################
### seed (testdata fixtures)
###############################################################################
#
# このプロジェクトはライブラリで DB を持たない。
# `make seed` は国税庁 Web-API から実レスポンス XML を取得して
# testdata/ 配下にゴールデンファイルとして保存し、オフラインで
# パーサ・enrichment 層のテストを書けるようにする。
#

.PHONY: seed
seed: ## Fetch real API responses into testdata/ (requires HOUJIN_APP_ID)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set. Export it before running seed."; \
		exit 1; \
	fi
	go run ./scripts/seed

.PHONY: seed-clean
seed-clean: ## Remove fetched testdata fixtures
	rm -rf testdata/seed

###############################################################################
### build / clean
###############################################################################

.PHONY: build
build: ## Verify all packages compile (output discarded; examples wouldn't pollute repo root)
	go build -o /dev/null ./...

FROM_DATE ?=
TO_DATE ?=
FORMAT ?= csv
OUTPUT ?= companies.csv
PREFECTURE ?=
LIMIT ?= 0
MANIFEST ?=

.PHONY: export-analysis
export-analysis: ## Export NTA updates as CSV/JSONL or Source Package JSON (FROM_DATE/TO_DATE required)
	@if [ -z "$$HOUJIN_APP_ID" ]; then \
		echo "ERROR: HOUJIN_APP_ID is not set."; \
		exit 1; \
	fi
	@if [ -z "$(FROM_DATE)" ] || [ -z "$(TO_DATE)" ]; then \
		echo "ERROR: FROM_DATE and TO_DATE are required."; \
		exit 1; \
	fi
	go run ./cmd/company-catalog \
		-from "$(FROM_DATE)" \
		-to "$(TO_DATE)" \
		-format "$(FORMAT)" \
		-output "$(OUTPUT)" \
		-prefecture "$(PREFECTURE)" \
		-limit "$(LIMIT)" \
		-manifest "$(MANIFEST)"

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: clean
clean: ## Clean caches and generated files
	go clean -testcache
	rm -f coverage.out
