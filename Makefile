export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

include .bingo/Variables.mk

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=-mod=vendor
export GO111MODULE=on

export REPORT_DIR?=$(CURDIR)/reports/$(shell date +%Y-%m-%d-%H-%M-%S)

export KUBECTL=$(shell command -v kubectl)

ES_CONTAINER_NAME=origin-logging-elasticsearch6
ES_IMAGE_TAG=quay.io/openshift/origin-logging-elasticsearch6:latest

all: lint bench-dev

lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

$(REPORT_DIR):
	@mkdir -p $(REPORT_DIR)

run-es:
	docker run -d --name $(ES_CONTAINER_NAME) \
		-p 9200:9200 -p 9300:9300 \
		-e "discovery.type=single-node" \
		$(ES_IMAGE_TAG)

bench-dev: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR) run-es
	@TARGET_ENV=development ./run.sh
.PHONY: bench-dev
