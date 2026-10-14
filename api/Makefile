BUILD_DIR ?= build
CRD_MODULE_PATH ?= $(shell go list -m)/${BUILD_DIR}/
COMPILER_TAG ?= latest
CONTAINER_ID ?= ""
DATAMODEL_LOCAL_PATH ?= $(realpath .)
BUCKET ?= nexus-template-downloads
DOCKER_REPO ?= $(shell cat nexus.yaml | grep dockerRepo | awk '{print $$2}' | tr -d '"')
DATAMODEL_DOCKER_REGISTRY ?= $(shell cat nexus.yaml | grep dockerRepo | awk '{print $$2}' | tr -d '"')
DATAMODEL_GROUP_NAME ?= $(shell cat nexus.yaml | grep groupName | awk '{print $$2}' | tr -d '"')
TAG ?= latest
NAME ?= $(shell cat nexus.yaml | grep groupName | awk '{print $$2}' | tr -d '"')
NEXUS_API_GW_K8S_NAMESPACE ?= "default"
DOCKER_BASE_IMAGE=gcr.io/nsx-sm/photon-kubectl:latest
DOCKER_NETWORK ?= nexus
MOUNTED_KUBECONFIG ?= /etc/config/kubeconfig

ifeq (${DATAMODEL_DOCKER_REGISTRY}, )
DATAMODEL_DOCKER_TAG=nexus/${DATAMODEL_GROUP_NAME}:${TAG}
else
DATAMODEL_DOCKER_TAG=${DATAMODEL_DOCKER_REGISTRY}/nexus/${DATAMODEL_GROUP_NAME}:${TAG}
endif

COMPILER_DOCKER_REGISTRY ?= ghcr.io/dilipmighty245
IMAGE_NAME ?= nexus/compiler
GOARCH ?= $(shell go env GOARCH)
COMPILER_NAME ?= ${IMAGE_NAME}/${GOARCH}
ifeq (${COMPILER_DOCKER_REGISTRY}, )
COMPILER_DOCKER_TAG = ${COMPILER_NAME}:${COMPILER_TAG}
else
COMPILER_DOCKER_TAG = ${COMPILER_DOCKER_REGISTRY}/${COMPILER_NAME}:${COMPILER_TAG}
endif

.PHONY: datamodel_build
datamodel_build:
	@echo "CRD and API Generated Output Directory: ${BUILD_DIR}"
	@echo "OPENAPISpec Generated Output Directory: ${BUILD_DIR}/crds/"
	rm -rf ${DATAMODEL_LOCAL_PATH}/build
	mkdir -p ${BUILD_DIR}
	if [ -z $(CONTAINER_ID) ]; then \
		docker run --pull=missing\
			--volume $(realpath .):/go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/datamodel/ \
			-v /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/datamodel/build/ \
			--volume $(realpath .)/build:/go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/generated/ \
			--volume ~/.ssh:/root/.ssh \
			--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
	        -e GOCACHE=/root/.cache/go-build \
			-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
			-e CONFIG_FILE=datamodel/nexus.yaml \
			${COMPILER_DOCKER_TAG}; \
	else \
		docker run --pull=missing\
			--volumes-from=$(CONTAINER_ID) \
			-e DATAMODEL_PATH=$(DATAMODEL_LOCAL_PATH) \
			-e GENERATED_OUTPUT_DIRECTORY=$(DATAMODEL_LOCAL_PATH)/build \
			-e CONFIG_FILE=${DATAMODEL_LOCAL_PATH}/nexus.yaml \
			-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
			${COMPILER_DOCKER_TAG}; \
	fi

.PHONY: docker.build
docker.build: datamodel_build
	docker build --build-arg DOCKER_BASE_IMAGE=${DOCKER_BASE_IMAGE} --build-arg IMAGE_NAME=${DATAMODEL_DOCKER_TAG} --build-arg NAME=${DATAMODEL_GROUP_NAME} -t ${DATAMODEL_DOCKER_TAG} . -f Dockerfile

.PHONY: docker.publish
docker.publish:
	docker push $(DATAMODEL_DOCKER_TAG)

.PHONY: dm.install
dm.install:
	docker run \
		--net ${DOCKER_NETWORK} \
		--pull=missing \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e NAME=${DATAMODEL_GROUP_NAME} \
		-e IMAGE=${DATAMODEL_DOCKER_TAG} \
		-e https_proxy= \
		-e http_proxy= \
		-e HTTPS_PROXY= \
		-e HTTP_PROXY= \
		${DATAMODEL_DOCKER_TAG}

.PHONY: dm.install.helm
dm.install.helm:
	cd helm-chart; helm install api -n ${NEXUS_API_GW_K8S_NAMESPACE} --set global.datamodelName=api --set image=${DATAMODEL_DOCKER_TAG} .

.PHONY: dm.uninstall.helm
dm.uninstall.helm:
	helm uninstall api
