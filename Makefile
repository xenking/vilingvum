lint: ## Run linter for the code
	fmt
	golangci-lint run ./... --fix --timeout 10m -v -c .golangci.yml

fmt: ## Format all the code
	gofumpt -l -w .
	gci -w -local tgbot/ .

generate: ## Generate the code
	gen-db-models
	fmt

gen-db-models: ## Generate the DB models
	sqlc -f ./database/sql/sqlc.yml generate

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
