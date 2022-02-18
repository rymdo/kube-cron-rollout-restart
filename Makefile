ifeq ($(shell uname -p), i386)
	HOST_ARCH = amd64
else ifeq ($(shell uname -p), x86_64)
	HOST_ARCH = amd64
else
	HOST_ARCH = arm64v8
endif

.PHONY: clean build run

default: clean build

clean:
	rm -rf dist

build:
	CONTAINER_NAME=$(shell basename $(CURDIR)) goreleaser release --snapshot --rm-dist
	docker images $(shell basename $(CURDIR))

run: clean build
	docker run -e USE_KUBECONFIG=true -v ~/.kube/config:/.kube/config --rm $(shell basename $(CURDIR)):latest-$(HOST_ARCH)