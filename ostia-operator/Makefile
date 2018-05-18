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
