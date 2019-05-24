.DEFAULT_GOAL := help

NAME=ostia-operator
NAMESPACE=quay.io/3scale
VERSION := latest
IMAGE ?= $(NAMESPACE)/$(NAME):$(VERSION)

.PHONY: all build generate test bash push unit integration clean_integration

export GO111MODULE=on

all: generate build

generate: ## Generate k8s resources
	@operator-sdk generate k8s

build: ## Build the operator
	@operator-sdk build $(IMAGE)

test: unit e2e ## Run all tests

bash: ## Start a container with bash
	docker run -it $(IMAGE) bash

push: ## Push to container registry
	docker push $(IMAGE)

vendor: vendor/modules.txt
vendor/modules.txt: go.sum
	go mod vendor -v

unit: ## Run unit tests
	go test ./...

e2e: ## Run operator-sdk e2e tests
	operator-sdk test local ./test/e2e --debug --go-test-flags "-v"

# Check http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
