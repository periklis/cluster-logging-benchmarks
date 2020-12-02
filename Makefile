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
ES_DEPLOYMENT_NAMESPACE?=openshift-logging

all: lint bench-dev

clean:
	@rm -rf tmp/*

lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

$(REPORT_DIR):
	@mkdir -p $(REPORT_DIR)

es-dirs:
	@mkdir -p ./tmp/elasticsearch/config
	@mkdir -p ./tmp/elasticsearch/configmap/config
	@mkdir -p ./tmp/elasticsearch/configmap/secrets
	@mkdir -p ./tmp/elasticsearch/data

run-es: gen-example-certs gen-es-config
	docker run -d --name $(ES_CONTAINER_NAME) \
		-p 9200:9200 -p 9300:9300 \
		-e "CLUSTER_NAME=elasticsearch" \
		-e "NAMESPACE=openshift-logging" \
		-e "INSTANCE_RAM=2G" \
		-v "$$(pwd)/tmp/elasticsearch/config:/etc/elasticsearch:rw" \
		-v "$$(pwd)/tmp/elasticsearch/data:/elasticsearch/persistent:rw" \
		-v "$$(pwd)/tmp/elasticsearch/configmap/secrets:/etc/elasticsearch//secret:rw" \
		-v "$$(pwd)/tmp/elasticsearch/configmap/config:/usr/share/java/elasticsearch/config:rw" \
		$(ES_IMAGE_TAG)

gen-es-config: es-dirs
	@cp ./hack/es-config/* ./tmp/elasticsearch/config
	@touch ./tmp/elasticsearch/configmap/config/test

gen-example-certs: es-dirs
	@rm -rf ./tmp/elasticsearch/configmap/secrets ||: \
	mkdir ./tmp/elasticsearch/configmap/secrets && \
	hack/cert_generation.sh ./tmp/elasticsearch/configmap/secrets $(DEPLOYMENT_NAMESPACE) elasticsearch
.PHONY: gen-example-certs

bench-dev: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR) run-es
	@TARGET_ENV=development ./run.sh
.PHONY: bench-dev
