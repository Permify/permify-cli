export

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build/compile the Permify service
	go build -o ./permctl ./cmd/permctl

.PHONY: goreleaser
goreleaser: ## Build goreleaser binaries
	goreleaser release --snapshot --clean

.PHONY: format
format: ## Auto-format the code
	gofumpt -l -w -extra .

.PHONY: run
run: ## Run cli
	go run ./cmd/permctl/*