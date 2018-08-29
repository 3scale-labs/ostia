NAME=ostia-operator
NAMESPACE=quay.io/3scale
VERSION := v0.1
IMAGE := $(NAMESPACE)/$(NAME):$(VERSION)

all: generate build

generate:
	@operator-sdk generate k8s

build:
	@operator-sdk build $(IMAGE)

test: build
	@docker run -t $(IMAGE) ostia-operator

bash:
	docker run -it $(IMAGE) bash

push:
	docker push $(IMAGE)

integration: clean_integration
	cd test && ./integration.sh ${OPENSHIFT_PUBLIC_HOSTNAME}

clean_integration: SHELL:=/bin/bash
clean_integration:
	if [[ $$(oc get projects -o custom-columns=:.metadata.name -l ostia-test=true --no-headers=true) != "" ]]; then \
	  oc delete project $$(oc get projects -o custom-columns=:.metadata.name -l ostia-test=true --no-headers=true); \
	fi
