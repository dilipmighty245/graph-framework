DOCKER_GO_PATH ?= /go
DOCKER_BUILD_MOUNT_DIR ?= ${DOCKER_GO_PATH}/src/github.com/dilipmighty245/graph-framework-for-microservices
NAMESPACE ?= default
CLI_DIR ?= cli
HOST_KUBECONFIG ?= ${HOME}/.kube/config
MOUNTED_KUBECONFIG ?= /etc/config/kubeconfig
RUNTIME_NAMESPACE ?= default
UNAME_S ?= $(shell uname -s)
CWD ?= $(shell pwd)
DATAMODEL_NAME ?= $(shell grep groupName ${DATAMODEL_DIR}/nexus.yaml | cut -f 2 -d" " | tr -d '"')
DATAMODEL_IMAGE_NAME ?= $(shell grep dockerRepo ${DATAMODEL_DIR}/nexus.yaml | cut -f 2 -d" ")
TAG ?= $(shell cat TAG | awk '{ print $1 }')
COMPILER_TAG ?= latest
CLUSTER_PORT ?= 9000
K8S_RUNTIME_PORT = $(shell echo $$(( ${CLUSTER_PORT} + 1 )))
LOG_LEVEL ?= ERROR
DOCKER_REGISTRY ?=
DOCKER_IMAGE_PREFIX = $(if $(strip $(DOCKER_REGISTRY)),$(patsubst %/,%,$(DOCKER_REGISTRY))/)
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
DOCKER_BUILDER_PLATFORM ?= linux/${ARCH}
ADMIN_DATAMODEL_DEFAULT_RUN_TAG ?= latest
API_GW_DEFAULT_RUN_TAG ?= latest
OPENAPI_GENERATOR_COMPONENT_NAME ?= openapi-generator

KIND_API_GW_PORT=$(shell expr ${CLUSTER_PORT} + 1)
API_GW_COMPONENT_NAME ?= api-gw
API_GW_DOCKER_IMAGE ?= ${DOCKER_IMAGE_PREFIX}nexus/${API_GW_COMPONENT_NAME}:${TAG}
API_GW_RUN_DOCKER_IMAGE ?= ${DOCKER_IMAGE_PREFIX}nexus/${API_GW_COMPONENT_NAME}:${API_GW_DEFAULT_RUN_TAG}
OPENAPI_GENERATOR_DOCKER_IMAGE?= ${DOCKER_IMAGE_PREFIX}nexus/${OPENAPI_GENERATOR_COMPONENT_NAME}:${TAG}

.PHONY: check.prereq.git
check.prereq.git:
	$(info Checking access to required git repositories ...)
	@git ls-remote git@github.com:kubernetes/code-generator.git > /dev/null

.PHONY: check.prereq.go
check.prereq.go:
	$(info Checking if you golang requirements are met...)
	@go version > /dev/null

.PHONY: check.prereq.docker
check.prereq.docker:
	$(info Checking if you docker requirements are met...)
	@docker ps > /dev/null

.PHONY: check.prereq
check.prereq: check.prereq.git check.prereq.go check.prereq.docker

.PHONY: check.env
check.env:
ifndef NEXUS_REPO_DIR
	$(error NEXUS_REPO_DIR is mandatory and should be set to the Nexus repo directory)
endif

.PHONY: check.repodir
check.repodir: check.env
ifneq (${NEXUS_REPO_DIR}, ${CWD})
	$(error NEXUS_REPO_DIR and current working directory does not match)
endif

.PHONY: check.tag
check.tag:
ifndef TAG
	$(error TAG is mandatory and should be set to the desired docker image tag)
endif

.PHONY: create.nexus.docker.network
create.nexus.docker.network:
	docker network create nexus | true

.PHONY: rm.nexus.docker.network
rm.nexus.docker.network:
	docker network rm nexus | true

.PHONY: cli.build.darwin
cli.build.darwin:
	docker run \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
		-e GOCACHE=/root/.cache/go-build \
		-w ${DOCKER_BUILD_MOUNT_DIR}/${CLI_DIR} \
		docker.io/library/golang:1.27.2-bookworm \
		/bin/bash -c "git config --global --add safe.directory '*' && go mod download && make build.darwin";

.PHONY: cli.build.linux
cli.build.linux:
	docker run \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		-w ${DOCKER_BUILD_MOUNT_DIR}/${CLI_DIR} \
		--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
		-e GOCACHE=/root/.cache/go-build \
		docker.io/library/golang:1.27.2-bookworm \
		/bin/bash -c "git config --global --add safe.directory '*' && go mod download && make build.linux";

.PHONY: cli.install.darwin
cli.install.darwin:
	cd cli; make install.darwin

.PHONY: cli.install.linux
cli.install.linux:
	cd cli; make install.linux

.PHONY: compiler.builder
compiler.builder:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${COMPILER_TAG} make docker.builder

.PHONY: compiler.builder.publish
compiler.builder.publish: check.registry
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${COMPILER_TAG} make docker.builder.publish

.PHONY: compiler.build
compiler.build:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} TAG=${TAG} make docker

.PHONY: compiler.build.publish
compiler.build.publish: check.registry
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} TAG=${TAG} make docker.publish

#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make api-gw.docker
# Tag and publish nexus api gw.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=letstest make api-gw.docker
#
.PHONY: api-gw.docker
api-gw.docker: check.tag
	docker build --build-arg API_GW_COMPONENT_NAME=${API_GW_COMPONENT_NAME} -t ${API_GW_DOCKER_IMAGE} -f api-gw/Dockerfile .

#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make api-gw.docker.kubeconfig
# Tag and publish nexus api gw for execution on K8s cluster.
# This forces the the api-gw to be built for linux/amd64 arch.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=letstest make api-gw.docker.kubeconfig
#
.PHONY: api-gw.docker.kubeconfig
api-gw.docker.kubeconfig:
	docker run \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		-w ${DOCKER_BUILD_MOUNT_DIR}/${API_GW_COMPONENT_NAME} \
		docker.io/library/golang:1.27.2-bookworm \
		/bin/bash -c "go mod download && GOOS=linux GOARCH=amd64 go build -buildvcs=false -o bin/${API_GW_COMPONENT_NAME}";
	docker build --platform ${DOCKER_BUILDER_PLATFORM} -t ${API_GW_DOCKER_IMAGE} -f api-gw/Dockerfile .

.PHONY: api-gw.docker.publish
api-gw.docker.publish: check.registry
	docker push ${API_GW_DOCKER_IMAGE}

.PHONY: k0s.install
k0s.install:
	$(realpath .)/nexus-runtime-manifests/k0s/run_k0s.sh

.PHONY: k0s.uninstall
k0s.uninstall:
	$(realpath .)/nexus-runtime-manifests/k0s/stop_k0s.sh

.PHONY: k8s.install
k8s.install:
	$(realpath .)/nexus-runtime-manifests/k8s/run_k8s.sh -p ${CLUSTER_PORT}

.PHONY: k8s.uninstall
k8s.uninstall:
	$(realpath .)/nexus-runtime-manifests/k8s/run_k8s.sh -p ${CLUSTER_PORT} -d

.PHONY: dm.install_init
dm.install_init:
	docker run \
		--net host \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		-w ${DOCKER_BUILD_MOUNT_DIR}/nexus-runtime-manifests/datamodel-install \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		alpine/helm \
		upgrade --install datamodel-install-scripts . --set global.namespace=${RUNTIME_NAMESPACE} --kubeconfig ${MOUNTED_KUBECONFIG}

.PHONY: api.build
api.build:
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} COMPILER_TAG=${COMPILER_TAG} TAG=${TAG} make docker.build

.PHONY: api.install
api.install:
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${ADMIN_DATAMODEL_DEFAULT_RUN_TAG} MOUNTED_KUBECONFIG=${MOUNTED_KUBECONFIG} HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k8s/kubeconfig DOCKER_NETWORK=nexus make dm.install

.PHONY: api.install.k0s
api.install.k0s:
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${ADMIN_DATAMODEL_DEFAULT_RUN_TAG} HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k0s/.kubeconfig MOUNTED_KUBECONFIG=${MOUNTED_KUBECONFIG} DOCKER_NETWORK=nexus make dm.install

.PHONY: api.install.kind
api.install.kind:
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${ADMIN_DATAMODEL_DEFAULT_RUN_TAG} HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/kubeconfig MOUNTED_KUBECONFIG=${MOUNTED_KUBECONFIG} DOCKER_NETWORK=kind make dm.install

.PHONY: api-gw.run
api-gw.run: api-gw.stop
	docker run -d \
		--name=nexus-api-gw \
                --network nexus \
		--restart=always \
		--pull=missing \
		-p 8082:8082 \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		--mount type=bind,source=$(realpath .)/${API_GW_COMPONENT_NAME}/deploy/config/api-gw-config.yaml,target=/api-gw-config.yaml,readonly \
		-e APIGWCONFIG=/api-gw-config.yaml \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e LOG_LEVEL=DEBUG \
		-e KUBEAPI_ENDPOINT="http://k8s-proxy:${CLUSTER_PORT}" \
		-e https_proxy='' \
		-e http_proxy='' \
		-e HTTPS_PROXY='' \
		-e HTTP_PROXY='' \
		${API_GW_RUN_DOCKER_IMAGE}

.PHONY: api-gw.stop
api-gw.stop:
	docker rm -f nexus-api-gw > /dev/null || true

.PHONY: api-gw.kind.run
api-gw.kind.run:
ifndef CLUSTER_NAME
	$(error CLUSTER_NAME is mandatory)
endif
ifndef CLUSTER_PORT
	$(error CLUSTER_PORT is mandatory)
endif
	docker rm -f nexus-api-gw-${CLUSTER_NAME} > /dev/null || true
	docker run -d \
		--name=nexus-api-gw-${CLUSTER_NAME} \
		--rm \
		--network kind \
		--pull=missing \
		-p ${KIND_API_GW_PORT}:${KIND_API_GW_PORT} \
		--mount type=bind,source=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/kubeconfig,target=${MOUNTED_KUBECONFIG},readonly \
		--mount type=bind,source=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/api-gw-config.yaml,target=/api-gw-config.yaml,readonly \
		-e APIGWCONFIG=/api-gw-config.yaml \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e KUBEAPI_ENDPOINT="http://k8s-proxy-${CLUSTER_NAME}:${CLUSTER_PORT}" \
		-e LOG_LEVEL=${LOG_LEVEL} \
		${API_GW_RUN_DOCKER_IMAGE}

.PHONY: api-gw.k8s.run
api-gw.k8s.run: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k8s/kubeconfig
api-gw.k8s.run:
ifeq (${UNAME_S}, Linux)
	docker rm -f nexus-api-gw-${CLUSTER_NAME} > /dev/null || true
	docker run -d \
		--name=nexus-api-gw \
		--rm \
		--entrypoint /bin/bash \
		--network host \
		--pull=missing \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		--mount type=bind,source=$(realpath .)/${API_GW_COMPONENT_NAME}/deploy/config/api-gw-config.yaml,target=/api-gw-config.yaml,readonly \
		-e APIGWCONFIG=/api-gw-config.yaml \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e KUBEAPI_ENDPOINT="http://127.0.0.1:${CLUSTER_PORT}" \
		-e LOG_LEVEL=${LOG_LEVEL} \
		${API_GW_COMPONENT_NAME}:${TAG} -c "api-gw --metrics-bind-address :8083"
else
	APIGWCONFIG=$(realpath .)/${API_GW_COMPONENT_NAME}/deploy/config/api-gw-config.yaml KUBECONFIG=${HOST_KUBECONFIG} $(realpath .)/${API_GW_COMPONENT_NAME}/bin/${API_GW_COMPONENT_NAME}
endif

.PHONY: api-gw.kind.stop
api-gw.kind.stop:
	docker rm -f nexus-api-gw-${CLUSTER_NAME} > /dev/null || true

.PHONY: openapi.generator.docker
openapi.generator.docker:
	docker build -t ${OPENAPI_GENERATOR_DOCKER_IMAGE} -f api-gw/Dockerfile.openapi-generator .

.PHONY: openapi.generator.docker.publish
openapi.generator.docker.publish:
	docker push ${OPENAPI_GENERATOR_DOCKER_IMAGE}

.PHONY: runtime.build
runtime.build: api.build api-gw.docker

.PHONY: runtime.clean
runtime.clean:
	rm -rf api/build/*

.PHONY: runtime.install.k0s 
runtime.install.k0s: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k0s/.kubeconfig
runtime.install.k0s: check.repodir create.nexus.docker.network k0s.install api.install.k0s api-gw.run
	$(info )
	$(info ====================================================)
	$(info To access runtime, you can execute kubectl as:)
	$(info     kubectl -s localhost:8082 ...)
	$(info )
	$(info )
	$(info To access nexus api gateway using kubeconfig, export:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info )
	$(info ====================================================)

.PHONY: runtime.uninstall.k0s
runtime.uninstall.k0s: k0s.uninstall api-gw.stop rm.nexus.docker.network
	$(info )
	$(info ====================================================)
	$(info Runtime is now uninstalled)
	$(info ====================================================)

.PHONY: kind.install
kind.install:
	$(realpath .)/nexus-runtime-manifests/kind/run_kind.sh -n ${CLUSTER_NAME} -p ${CLUSTER_PORT}

.PHONY: kind.uninstall
kind.uninstall:
	$(realpath .)/nexus-runtime-manifests/kind/run_kind.sh -n ${CLUSTER_NAME} -d

.PHONY: check.kind
check.kind:
	$(info Checking if kind requirements are met...)
	@kind version > /dev/null

.PHONY: runtime.uninstall.kind
runtime.uninstall.kind: check.kind kind.uninstall api-gw.kind.stop
	$(info )
	$(info ====================================================)
	$(info Runtime is now uninstalled)
	$(info ====================================================)

.PHONY: runtime.install.kind
runtime.install.kind: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/kubeconfig
runtime.install.kind: check.kind check.repodir kind.install api.install.kind api-gw.kind.run
	$(info )
	$(info ====================================================)
	$(info To access runtime, you can execute kubectl as:)
	$(info     kubectl -s localhost:${K8S_RUNTIME_PORT} ...)
	$(info )
	$(info To access nexus api gateway using kubeconfig, export:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info     export DOCKER_NETWORK=kind)
	$(info )
	$(info ====================================================)

.PHONY: runtime.install
runtime.install: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k8s/kubeconfig
runtime.install: HOST_KUBECONFIG_LOCAL=$(realpath .)/nexus-runtime-manifests/k8s/kubeconfig.local
runtime.install: runtime.uninstall create.nexus.docker.network k8s.install api.install api-gw.run
	$(info )
	$(info ====================================================)
	$(info To access runtime, you can execute kubectl as:)
	$(info     kubectl -s localhost:${CLUSTER_PORT} ...)
	$(info )
	$(info )
	$(info For command line  access nexus api gateway using kubeconfig, export:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info )
	$(info For programmatic access to nexus runtime from a local process running on localhost, export:)
	$(info     export HOST_KUBECONFIG_LOCAL=${HOST_KUBECONFIG_LOCAL})
	$(info )
	$(info ====================================================)

.PHONY: runtime.uninstall
runtime.uninstall: k8s.uninstall api-gw.stop
	$(info )
	$(info ====================================================)
	$(info Runtime is now uninstalled)
	$(info ====================================================)

.PHONY: dm.check-datamodel-dir
dm.check-datamodel-dir:
ifndef DATAMODEL_DIR
	$(error DATAMODEL_DIR is mandatory)
endif

.PHONY: dm.install
dm.install: dm.check-datamodel-dir
	docker run \
		--net host \
		--pull=missing \
		--volume ${DATAMODEL_DIR}/build/crds:/crds \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		--mount type=bind,source=$(realpath .)/nexus-runtime-manifests/datamodel-install/datamodel_installer.sh,target=/datamodel_installer.sh,readonly \
		--entrypoint /datamodel_installer.sh \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e NAME=${DATAMODEL_NAME} \
		-e IMAGE=${DATAMODEL_IMAGE_NAME} \
		bitnami/kubectl

.PHONY: runtime.k0s.kubeconfig.export
runtime.k0s.kubeconfig.export: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/k0s/.kubeconfig
runtime.k0s.kubeconfig.export:
	$(info )
	$(info Execute the below export statement on your shell:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info )
	@echo > /dev/null

.PHONY: runtime.kind.kubeconfig.export
runtime.kind.kubeconfig.export: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/kubeconfig
runtime.kind.kubeconfig.export:
ifndef CLUSTER_NAME
	$(error CLUSTER_NAME is mandatory)
else
	$(info )
	$(info Execute the below export statement on your shell:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info     export DOCKER_NETWORK=kind
	$(info )
endif
	@echo > /dev/null

#
# Usage: TAG=<tag-value> make core.runtime.build
# Build artifacts that mke up nexus runtime core.
#
# Example: TAG=latest make core.runtime.build
#
.PHONY: core.runtime.build
core.runtime.build: api.build api-gw.docker.kubeconfig

#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make core.runtime.release
# Publish core artifiacats to container registry.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=latest make core.runtime.release
#
.PHONY: core.runtime.release
core.runtime.release: check.registry
	docker push ${API_GW_DOCKER_IMAGE}
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${TAG} make docker.publish
#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make core.runtime.download
# Download all artifacts that make up the nexus runtime.
# TAG provides a way to download custom tagged artifacts.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=latest make core.runtime.download
#
.PHONY: core.runtime.download
core.runtime.download: check.registry
	docker pull ${DOCKER_IMAGE_PREFIX}nexus/api-gw:${TAG}
	docker pull ${DOCKER_IMAGE_PREFIX}nexus/admin.nexus.com:${TAG}
	docker pull ${DOCKER_IMAGE_PREFIX}nexus/debugtools:latest
	docker pull ${DOCKER_IMAGE_PREFIX}nexus/nexus-kube-apiserver:latest
	docker pull ${DOCKER_IMAGE_PREFIX}nexus/nexus-etcd-kubectl:latest

#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make core.runtime.artifacts.push
# Tag and publish all the images needed by nexus rutnime.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=letstest make core.runtime.artifacts.push
#
.PHONY: core.runtime.artifacts.push
core.runtime.artifacts.push: check.registry
	docker tag nexus/api-gw:${TAG} ${API_GW_DOCKER_IMAGE}
	docker push ${API_GW_DOCKER_IMAGE}
	docker tag nexus/admin.nexus.com:${TAG} ${DOCKER_IMAGE_PREFIX}nexus/admin.nexus.com:${TAG}
	docker push ${DOCKER_IMAGE_PREFIX}nexus/admin.nexus.com:${TAG}
	docker tag nexus/debugtools:latest ${DOCKER_IMAGE_PREFIX}nexus/debugtools:${TAG}
	docker push ${DOCKER_IMAGE_PREFIX}nexus/debugtools:${TAG}
	docker tag nexus/nexus-kube-apiserver:latest ${DOCKER_IMAGE_PREFIX}nexus/nexus-kube-apiserver:${TAG}
	docker push ${DOCKER_IMAGE_PREFIX}nexus/nexus-kube-apiserver:${TAG}
	docker tag nexus/nexus-etcd-kubectl:latest ${DOCKER_IMAGE_PREFIX}nexus/nexus-etcd-kubectl:${TAG}
	docker push ${DOCKER_IMAGE_PREFIX}nexus/nexus-etcd-kubectl:${TAG}

#
# Usage: DOCKER_REGISTRY=<registry> TAG=<tag-value> make core.install.kubeconfig
# Tag and publish all the images needed by nexus rutnime.
#
# Example: DOCKER_REGISTRY=<your-registry> TAG=letstest make core.install.kubeconfig
#
.PHONY: core.install.kubeconfig
core.install.kubeconfig:
ifndef TAG
	$(error TAG is mandatory)
endif
ifndef DOCKER_REGISTRY
	$(error DOCKER_REGISTRY is mandatory)
endif
	cd nexus-runtime-manifests/helm-charts/core; helm install core -n ${NAMESPACE} --set imageTag=${TAG} --set global.resources.kubeapiserver.cpu=6 --set global.tainted=true --set imageRegistry=${DOCKER_REGISTRY} .
	cd api; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${TAG} make dm.install.helm

.PHONY: core.uninstall.kubeconfig
core.uninstall.kubeconfig:
	helm uninstall -n ${NAMESPACE} core | true
	helm uninstall -n ${NAMESPACE} api | true

.PHONY: demo.install.kubeconfig
demo.install.kubeconfig: core.install.kubeconfig
	cd demo/edge-power-scheduler/datamodel; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${TAG} make dm.install.helm
	cd demo/edge-power-scheduler/helm-chart; helm install epm-app --set imageTag=${TAG} --set global.tainted=true --set imageRegistry=${DOCKER_REGISTRY} --namespace ${NAMESPACE}  .
	kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

.PHONY: demo.uninstall.kubeconfig
demo.uninstall.kubeconfig: core.uninstall.kubeconfig
	helm uninstall -n ${NAMESPACE} epm | true
	helm uninstall -n ${NAMESPACE}  epm-app | true
	kubectl delete -n ${NAMESPACE} pvc data-nexus-etcd-0
	kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

.PHONY: traffic.demo.install.kubeconfig
traffic.demo.install.kubeconfig: core.install.kubeconfig
	cd demo/traffic-lights/datamodel; DATAMODEL_DOCKER_REGISTRY=${DOCKER_REGISTRY} TAG=${TAG} make dm.install.helm
	cd demo/traffic-lights/helm-chart; helm install trafficlightapp --set imageTag=${TAG} --set global.tainted=true --set imageRegistry=${DOCKER_REGISTRY} --namespace ${NAMESPACE}  .
	kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

.PHONY: traffic.demo.uninstall.kubeconfig
traffic.demo.uninstall.kubeconfig: core.uninstall.kubeconfig
	helm uninstall -n ${NAMESPACE} trafficlight | true
	helm uninstall -n ${NAMESPACE}  trafficlightapp | true
	kubectl delete -n ${NAMESPACE} pvc data-nexus-etcd-0
	kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Cross-module checks; see docs/testing.md for prerequisites and coverage.
.PHONY: clean clean.preview clean.all
clean:
	@CLUSTER_NAME="$(CLUSTER_NAME)" python3 scripts/clean-framework.py

clean.preview:
	@CLUSTER_NAME="$(CLUSTER_NAME)" python3 scripts/clean-framework.py --dry-run

# Explicitly includes unrelated local Kind clusters. Preview with the script first.
clean.all:
	@CLUSTER_NAME="$(CLUSTER_NAME)" python3 scripts/clean-framework.py --all-kind-clusters

.PHONY: test.framework.builders test.framework test.framework.compiler test.framework.gateway test.framework.all test.framework.e2e test.framework.runtime test.framework.stability test.framework.race test.framework.help
test.framework:
	@bash scripts/test-framework.sh regression

test.framework.gateway:
	@bash scripts/test-framework.sh gateway

test.framework.compiler:
	@bash scripts/test-framework.sh compiler

test.framework.all: test.framework.stability

test.framework.stability:
	@FRAMEWORK_TEST_RUNS=$${FRAMEWORK_TEST_RUNS:-3} FRAMEWORK_TEST_RACE=$${FRAMEWORK_TEST_RACE:-true} bash scripts/test-framework.sh all

test.framework.race:
	@FRAMEWORK_TEST_RACE=true bash scripts/test-framework.sh regression

test.framework.builders:
	@bash scripts/test-framework-builders.sh

test.framework.e2e:
	@bash scripts/test-framework-e2e.sh

test.framework.runtime:
	@E2E_COMPONENTS=resources E2E_RUNTIME_COMPATIBILITY=true bash scripts/test-framework-e2e.sh

test.framework.help:
	@printf '%s\n' \
	  'make test.framework          Run complete source suites across framework modules' \
	  'make test.framework.gateway  Run gateway and controller suites; generated admin API required' \
	  'make test.framework.compiler Run every compiler command and package suite' \
	  'make test.framework.all      Run stability: source suites, Kind E2E, and both runtime charts (3 passes)' \
	  'make test.framework.builders Build public Go builders and run detailed Kind checks' \
	  'make test.framework.e2e      Generate artifacts and test resources, gateway, and GraphQL in Kind' \
	  'make test.framework.runtime  Generate a fresh SDK and test both runtime charts in Kind' \
	  'make test.framework.stability Repeat all checks with race detection (default: 3 passes)' \
	  'make test.framework.race     Run source suites with the race detector' \
	  'See docs/testing.md for host tools, optional compiler images, and logs.'

.PHONY: check.registry
check.registry:
	@test -n "${DOCKER_REGISTRY}" || { echo "Set DOCKER_REGISTRY explicitly for image download or publication." >&2; exit 2; }
