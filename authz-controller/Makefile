DEBUG ?= FALSE

GO_PROJECT_NAME ?= authz-controller
BUCKET_NAME ?= nexus-template-downloads

ECR_DOCKER_REGISTRY ?=
DOCKER_IMAGE_PREFIX = $(if $(strip $(DOCKER_REGISTRY)),$(patsubst %/,%,$(DOCKER_REGISTRY))/)
DOCKER_REGISTRY ?=
DOCKER_IMAGE_PREFIX = $(if $(strip $(DOCKER_REGISTRY)),$(patsubst %/,%,$(DOCKER_REGISTRY))/)
IMAGE_NAME ?= authz
TAG ?= $(shell git rev-parse --verify HEAD)

BUILDER_NAME ?= ${IMAGE_NAME}-builder
BUILDER_TAG := $(shell md5sum builder/Dockerfile | awk '{ print $1 }' | head -c 8)

PKG_NAME ?= /go/src/github.com/dilipmighty245/graph-framework-for-microservices/${GO_PROJECT_NAME}

ifeq ($(CONTAINER_ID),)
define run_in_container
  docker run \
  --volume $(realpath .)/..:/go/src/github.com/dilipmighty245/graph-framework-for-microservices/ \
  --workdir ${PKG_NAME} \
  "${BUILDER_NAME}:${BUILDER_TAG}" /bin/bash -c "make cred_setup && ${1}"
endef
else
define run_in_container
 docker run \
 --volumes-from ${CONTAINER_ID} \
 --workdir ${PKG_NAME} \
 "${BUILDER_NAME}:${BUILDER_TAG}" /bin/bash -c "make cred_setup && ${1}"
endef
endif

%.image.exists:
	@docker inspect $* >/dev/null 2>&1 || \
		(echo "Image $* does not exist. Use 'make docker.builder'." && false)

.PHONY: docker.builder
docker.builder:
	docker build --no-cache -t ${BUILDER_NAME}:${BUILDER_TAG} builder/

.PHONY: build
build:
	cd cmd/authz-controller && \
		CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" .

.PHONY: build_in_container
build_in_container: ${BUILDER_NAME}\:${BUILDER_TAG}.image.exists
	$(call run_in_container,make build)

.PHONY: tools
tools:
	go install golang.org/x/tools/cmd/goimports@v0.51.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
	go install github.com/onsi/ginkgo/ginkgo@v1.16.5

.PHONY: unit-test
unit-test:
	CGO_ENABLED=0 GOOS=linux go test -cover ./...

.PHONY: race-unit-test
race-unit-test:
	go test -race -cover ./...

.PHONY: test-fmt
test-fmt:
	test -z $$(goimports -w -l cmd pkg)

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test: test-fmt vet lint race-unit-test

.PHONY: test_in_container
test_in_container: ${BUILDER_NAME}\:${BUILDER_TAG}.image.exists
	$(call run_in_container, make test)

.PHONY: show-image-name
show-image-name:
	@echo ${DOCKER_IMAGE_PREFIX}${IMAGE_NAME}:${TAG}

.PHONY: docker
docker: build
	docker build --no-cache \
		--build-arg BUILDER_TAG=${BUILDER_TAG} \
		-t ${IMAGE_NAME}:${TAG} .

.PHONY: publish
publish: require_registry
	docker tag ${IMAGE_NAME}:${TAG} ${DOCKER_IMAGE_PREFIX}${IMAGE_NAME}:${TAG}
	docker push ${DOCKER_IMAGE_PREFIX}${IMAGE_NAME}:${TAG};

.PHONY: publish.ecr
publish.ecr: require_ecr_registry
	docker tag ${IMAGE_NAME}:${TAG} ${ECR_DOCKER_REGISTRY}/${IMAGE_NAME}:${TAG}
	docker push ${ECR_DOCKER_REGISTRY}/${IMAGE_NAME}:${TAG};

.PHONY: download_builder_image
download_builder_image: require_ecr_registry
	docker pull ${ECR_DOCKER_REGISTRY}/${BUILDER_NAME}:${BUILDER_TAG}
	docker tag ${ECR_DOCKER_REGISTRY}/${BUILDER_NAME}:${BUILDER_TAG} ${BUILDER_NAME}:${BUILDER_TAG}

.PHONY: publish_builder_image
publish_builder_image: require_ecr_registry
	docker tag ${BUILDER_NAME}:${BUILDER_TAG} ${ECR_DOCKER_REGISTRY}/${BUILDER_NAME}:${BUILDER_TAG}
	docker push ${ECR_DOCKER_REGISTRY}/${BUILDER_NAME}:${BUILDER_TAG}

.PHONY: coverage
coverage:
	go test -json -coverprofile=coverage.out -coverpkg=./... ./... | tee report.json ;

build_template:
	tar -czvf authz-controller-manifests.tar manifests/*

publish_template: build_template
	gsutil cp authz-controller-manifests.tar gs://${BUCKET_NAME}/${TAG}/

cred_setup:
	go mod download

.PHONY: require_ecr_registry
require_ecr_registry:
	@test -n "${ECR_DOCKER_REGISTRY}" || { echo 'Set ECR_DOCKER_REGISTRY explicitly for ECR operations.' >&2; exit 2; }

.PHONY: require_registry
require_registry:
	@test -n "${DOCKER_REGISTRY}" || { echo "Set DOCKER_REGISTRY explicitly for image publication." >&2; exit 2; }
