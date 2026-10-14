UNAME := $(shell uname)
NAMESPACE ?= default
BUILD_DIR ?= build
DATAMODEL ?= "datamodel"
OUTPUT_DIR ?= "${BUILD_DIR}/${DATAMODEL}/output"
DATAMODEL_DIR ?= "./nexus"
APP_DIR ?= ".."
PID=$(shell ps -ef | grep $$(cat /tmp/portforward.pid) | grep -v grep | wc -l| tr -d ' ')
FILE ?= ""
SHELL=bash

.PHONY: runtime_install
runtime_install:
	kubectl create namespace ${NAMESPACE} --dry-run -o yaml | kubectl apply -f -
	mkdir -p /tmp/ && \
	cd /tmp  && \
	curl https://storage.googleapis.com/nexus-template-downloads/runtime-manifests.tar -o manifests.tar && \
	tar -xvf manifests.tar && \
	kubectl apply -f runtime-manifests/crds/ -n ${NAMESPACE} &&\
	kubectl apply -f runtime-manifests/deployment/ -n ${NAMESPACE};
	$(MAKE) post-install-checks

.SILENT:
post-install-checks:
	CURRENT=0
	while [[ -z `kubectl get pods -n ${NAMESPACE} -lapp=etcd 2>/dev/null` ]] ; do \
		CURRENT=$$(( CURRENT + 1 )) ;\
		TRIES_LEFT=$$(( 20 - CURRENT )) ;\
		if [[ $$TRIES_LEFT -ne 0 ]] ; then \
			sleep 6 ;\
		else \
			echo "etcd pod not created"; \
			exit 1 ; \
		fi \
	done ; \
	kubectl wait pods -lapp=etcd -n ${NAMESPACE} --for=condition=ready --timeout=300s || { "etcd pod could not be started"; exit 1; }
	kubectl wait pods -lapp=kube-apiserver -n ${NAMESPACE} --for=condition=ready --timeout=120s || { "k8sapiserver pod could not be started"; exit 1; }
	kubectl wait pods -lname=kube-controllermanager -n ${NAMESPACE} --for=condition=ready --timeout=120s || { "k8scontrollermanager pod could not be started"; exit 1; }

.SILENT:
start-portforward: check-portforward
	( kubectl port-forward svc/nexus-proxy-container -n ${NAMESPACE} 45192:80 & echo $$! > /tmp/portforward.pid )
	CURRENT=0
	while [[ `curl -s -o /dev/null -w "%{http_code}" http://localhost:45192/api/v1/namespaces 2>/dev/null` -ne 200  ]] ; do \
		CURRENT=$$(( CURRENT + 1 )) ;\
		TRIES_LEFT=$$(( 10 - CURRENT )) ;\
		if [[ $$TRIES_LEFT -ne 0 ]] ; then \
			sleep 5 ;\
		else \
			echo "portforwarding failed due to service not reachable"; \
			$(MAKE) kill-portforward ;\
			exit 1 ; \
		fi \
	done

.SILENT:
check-portforward:
	@if [ -z `which telnet` ]; then \
		echo "please install telnet for mac : brew install telnet" ;\
		exit 1 ;\
	fi
	@if [ -z `kubectl get ns ${NAMESPACE} --no-headers 2>/dev/null` ]; then \
		echo "Please install runtime on namespace" ;\
		exit 1 ;\
	fi
	@if echo -e '\x1dclose\x0d' | telnet localhost 45192; then\
		echo "could not portforward to 45192 port - port already open"; exit 1; \
	else \
		echo "Proceeding to start portforwarding.." ;\
	fi

.SILENT:
kill-portforward:
	@if test -s /tmp/portforward.pid ; then \
		{ kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid ;} ;\
	fi

check-file:
	@if [ "$(FILE)" = "" ]; then \
		{ echo 'please call as FILE="<filename> make datamodel_apply'; exit 1; }  ;\
	fi

.PHONY: datamodel_install
datamodel_install: start-portforward
	kubectl -s localhost:45192 apply -f ${DATAMODEL}/build/crds/ || { echo "could not install datamodel"; kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid; exit 1; } && \
	kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid;	

.PHONY: datamodel_apply
datamodel_apply: check-file start-portforward
	kubectl -s localhost:45192 apply -f $(FILE) -n default || { echo "could not apply datamodel crds"; kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid; exit 1; } && \
	{ kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid; } ;\
	

.PHONY: datamodel_build
datamodel_build:
	$(MAKE) -C $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../../api datamodel_build

.PHONY: app_init
app_init:
	nexus app init

.PHONY: app_run
app_run: kill-portforward start-portforward
	cd ${APP_DIR} ; \
	{ go get .  ; go mod vendor  ; go run main.go -host http://localhost:45192; } || { kill $(shell cat /tmp/portforward.pid); rm /tmp/portforward.pid; exit 1; } ; 
