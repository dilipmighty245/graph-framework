BUILD_DIR ?= build
CRD_MODULE_PATH ?= $(shell go list -m)/${BUILD_DIR}/
CONTAINER_ID ?= ""
DATAMODEL_LOCAL_PATH ?= $(realpath .)
BUCKET ?= nexus-template-downloads
COMPILER_TAG ?= latest
VERSION ?= $(shell git rev-parse --verify HEAD 2>/dev/null)
DATAMODEL_DOCKER_REGISTRY ?= $(shell cat nexus.yaml | grep dockerRepo | awk '{print $$2}' | tr -d '"')
DATAMODEL_GROUP_NAME ?= $(shell cat nexus.yaml | grep groupName | awk '{print $$2}' | tr -d '"')
NAMESPACE ?= "default"
HOST_KUBECONFIG ?= ${HOME}/.kube/config
MOUNTED_KUBECONFIG ?= /etc/config/kubeconfig
DATAMODEL_BASE_DOCKER_IMAGE=gcr.io/nsx-sm/photon-kubectl:latest
DOCKER_NETWORK ?= nexus

# values for nexus datamodel upgrade
FORCE ?= true

ifeq (,$(shell echo ${ARTIFACT_REPO}))
ARTIFACT_REPO = $(shell cat nexus.yaml | grep artifactRepo | awk '{print $$2}' | tr -d '"')
endif

ifneq (,$(shell echo ${PREV_SPEC_BRANCH}))
# backward compatability check is required.
FORCE=false

# check if ARTIFACT_REPO is configured
ifeq (,$(shell echo ${ARTIFACT_REPO}))
# Error and exit because user provided the branch and artifact repo details are not available
$(error Unable to determine ARTIFACT_REPO to validate backward compatibility with ${PREV_SPEC_BRANCH} branch)
endif

TARGET_REPO_NAME = $(shell echo ${ARTIFACT_REPO} | awk -F'/' '{print $$2}' | tr -d '"'| awk -F'.git' '{print $$1}')

ifeq (,$(shell echo $(shell realpath ${TARGET_REPO_NAME}/crds)))
ARTIFACT_REPO_GITCMD=$(shell echo git clone $(shell echo ${ARTIFACT_REPO}) --branch $(shell echo ${PREV_SPEC_BRANCH}) --single-branch $(shell echo ${TARGET_REPO_NAME}))
else
ARTIFACT_REPO_GITCMD=$(shell echo git -C $(shell realpath ${TARGET_REPO_NAME}) pull origin $(shell echo ${PREV_SPEC_BRANCH}))
endif

endif

ifeq (${DATAMODEL_DOCKER_REGISTRY}, )
DATAMODEL_DOCKER_TAG=${DATAMODEL_GROUP_NAME}:${TAG}
else
DATAMODEL_DOCKER_TAG=${DATAMODEL_DOCKER_REGISTRY}/${DATAMODEL_GROUP_NAME}:${TAG}
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

.PHONEY: datamode_build_local_template
datamode_build_local_template:
	mkdir -p ${BUILD_DIR}
	if [ ${FORCE} = false ] ; then \
            		bash -c "${ARTIFACT_REPO_GITCMD}";\
              		cp -rf ${TARGET_REPO_NAME}/crds ${BUILD_DIR} ;\
                	echo "Copying of ${TARGET_REPO_NAME}/crds to the destination directory ${BUILD_DIR}/crds"; \
    fi ;\

	docker run --pull=missing\
		--volume $(realpath .):/go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/datamodel/ \
		-v /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/datamodel/build/ \
		--volume $(realpath .)/build:/go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/generated/ \
		--volume $(realpath ../../..)/compiler/pkg/generator/template:/go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/generator/template \
		--volume ~/.ssh:/root/.ssh \
		--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
		-e GOCACHE=/root/.cache/go-build \
		-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
		-e CONFIG_FILE=datamodel/nexus.yaml \
		-e FORCE=$(FORCE) \
		${COMPILER_DOCKER_TAG} /bin/bash -c 'make generate_code && chown -R $(shell id -u):$(shell id -g) .;'; \


.PHONY: datamodel_build
datamodel_build:
	mkdir -p ${BUILD_DIR}
	@echo "FORCE: ${FORCE}"
	@echo "PREV_SPEC_BRANCH: ${PREV_SPEC_BRANCH}"
	@echo "ARTIFACT_REPO: ${ARTIFACT_REPO}"
	@echo "TARGET_REPO_NAME: ${TARGET_REPO_NAME}"
	@echo "ARTIFACT_REPO_GITCMD: ${ARTIFACT_REPO_GITCMD}"
	@echo "CONTAINER_ID: ${CONTAINER_ID}"

	if [ ${FORCE} = false ] ; then \
            		bash -c "${ARTIFACT_REPO_GITCMD}";\
              		cp -rf ${TARGET_REPO_NAME}/crds ${BUILD_DIR} ;\
                	echo "Copying of ${TARGET_REPO_NAME}/crds to the destination directory ${BUILD_DIR}/crds"; \
    fi ;\

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
			-e FORCE=$(FORCE) \
			${COMPILER_DOCKER_TAG} /bin/bash -c 'make generate_code && chown -R $(shell id -u):$(shell id -g) .;'; \
	else \
		docker run --pull=missing\
			--volumes-from=$(CONTAINER_ID) \
			-e DATAMODEL_PATH=$(DATAMODEL_LOCAL_PATH) \
			-e GENERATED_OUTPUT_DIRECTORY=$(DATAMODEL_LOCAL_PATH)/build \
			-e CONFIG_FILE=${DATAMODEL_LOCAL_PATH}/nexus.yaml \
			-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
			-e FORCE=${FORCE} \
			--user root:root \
			-e FORCE=$(FORCE) \
			${COMPILER_DOCKER_TAG}; \
	fi

#
# Usage: DATAMODEL_DOCKER_REGISTRY=<registry> TAG=<tag-value> make docker.build
# Tag and build datamodel docker image.
#
# Example: DATAMODEL_DOCKER_REGISTRY=822995803632.dkr.ecr.us-west-2.amazonaws.com TAG=letstest make docker.build
#
.PHONY: docker.build
docker.build: datamodel_build
ifeq ($(DATAMODEL_DOCKER_REGISTRY),)
$(error ERROR: DATAMODEL_DOCKER_REGISTRY is a mandatory argument. Configure it to point to container registry.)
endif
ifeq ($(TAG),)
$(error ERROR: TAG is a mandatory argument. It should contain datamodel container image tag.)
endif
	docker build --build-arg DOCKER_BASE_IMAGE=${DATAMODEL_BASE_DOCKER_IMAGE} --build-arg IMAGE_NAME=${DATAMODEL_DOCKER_TAG} --build-arg NAME=${DATAMODEL_GROUP_NAME} -t ${DATAMODEL_DOCKER_TAG} . -f Dockerfile

#
# Usage: DATAMODEL_DOCKER_REGISTRY=<registry> TAG=<tag-value> make docker.publish
# Tag and publish datamodel docker image.
#
# Example: DATAMODEL_DOCKER_REGISTRY=822995803632.dkr.ecr.us-west-2.amazonaws.com TAG=letstest make docker.publish
#
.PHONY: docker.publish
docker.publish:
	docker push $(DATAMODEL_DOCKER_TAG)

.PHONY: dm.install
dm.install:
	docker run \
	    -u root \
		--net ${DOCKER_NETWORK} \
		--pull=missing \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e NAME=${DATAMODEL_GROUP_NAME} \
		-e IMAGE=IMAGE \
		${DATAMODEL_DOCKER_TAG}
#
# Usage: DATAMODEL_DOCKER_REGISTRY=<registry> TAG=<tag-value> make dm.install.helm
# Install datamodel using helm
#
# Example: DATAMODEL_DOCKER_REGISTRY=822995803632.dkr.ecr.us-west-2.amazonaws.com TAG=letstest make dm.install.helm
#
.PHONY: dm.install.helm
dm.install.helm:
ifeq ($(DATAMODEL_DOCKER_REGISTRY),)
$(info ERROR: DATAMODEL_DOCKER_REGISTRY is a mandatory argument. Configure it to point to container registry.)
endif
ifeq ($(TAG),)
$(info ERROR: TAG is a mandatory argument. It should contain datamodel container image tag.)
endif
	cd helm-chart; helm install epm -n ${NAMESPACE} --set global.datamodelName=epm --set image=${DATAMODEL_DOCKER_TAG} .

.PHONY: dm.uninstall.helm
dm.uninstall.helm:
	cd helm-chart; helm uninstall epm
