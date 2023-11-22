.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command to run in "$(APP_NAME)":"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo

.PHONY: test-unit
test-unit: ## Run unit tests
	go test -v -race -short  ./... -coverprofile cover.out
	@go tool cover -func cover.out | grep total | grep -Eo '[0-9]+\.[0-9]+' | xargs -I'{}' echo total '{}'%

.PHONY: coverage
coverage: ## Show coverage
	@go tool cover -html=cover.out

