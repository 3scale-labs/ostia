.DEFAULT_GOAL := help

NAME=ostia-operator
NAMESPACE=quay.io/3scale
VERSION := v0.1
IMAGE ?= $(NAMESPACE)/$(NAME):$(VERSION)

.PHONY: all build generate test bash push unit integration clean_integration

all: generate build

generate: ## Generate k8s resources
	@operator-sdk generate k8s

build: ## Build the operator
	@operator-sdk build $(IMAGE)

test: unit integration ## Run all tests

bash: ## Start a container with bash
	docker run -it $(IMAGE) bash

push: ## Push to container registry
	docker push $(IMAGE)

unit: ## Run unit tests
	go test ./...

integration: clean_integration ## Run integration tests
	cd test && IMAGE="$(IMAGE)" ./integration.sh ${OPENSHIFT_PUBLIC_HOSTNAME}

clean_integration: SHELL:=/bin/bash
clean_integration:
	if [[ $$(oc get projects -o custom-columns=:.metadata.name -l ostia-test=true --no-headers=true) != "" ]]; then \
	  oc delete project $$(oc get projects -o custom-columns=:.metadata.name -l ostia-test=true --no-headers=true); \
	fi

# Check http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
