.PHONY: default
default: help

.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## run tests
	go test -p 1 -covermode=count -coverprofile=coverage.out ./...

.PHONY: test-cover
test-cover: test ## run tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: generate
generate: ## run 'go generate' for all packages
	go generate ./...