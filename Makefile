.PHONY: help test tidy vet fmt hooks check modules

MODULES := errors apiresponse logging observability lifecycle mongodb

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

modules: ## List workspace modules
	@for m in $(MODULES); do echo $$m; done

test: ## Run tests in every module
	@failed=0; \
	for m in $(MODULES); do \
		echo "==> test $$m"; \
		(cd $$m && go test -count=1 ./...) || failed=1; \
	done; \
	if [ $$failed -ne 0 ]; then exit 1; fi

tidy: ## Run go mod tidy in every module
	@for m in $(MODULES); do \
		echo "==> tidy $$m"; \
		(cd $$m && go mod tidy); \
	done

vet: ## Run go vet in every module
	@failed=0; \
	for m in $(MODULES); do \
		echo "==> vet $$m"; \
		(cd $$m && go vet ./...) || failed=1; \
	done; \
	if [ $$failed -ne 0 ]; then exit 1; fi

fmt: ## Format Go sources in every module (gofmt)
	@for m in $(MODULES); do \
		echo "==> fmt $$m"; \
		(cd $$m && gofmt -w .); \
	done

hooks: ## Enable repo git hooks (runs make fmt on commit)
	git config core.hooksPath .githooks
	@echo "core.hooksPath set to .githooks"

check: vet test ## Run vet + tests across all modules
