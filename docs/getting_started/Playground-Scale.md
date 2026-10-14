# Playground - Nexus Scale / Edge Power Scheduler

Welcome to the Nexus playground to demonstrate its scale.

This tutorial will walk through a workflow that will:

* build Nexus Runtime
* build datamodel for demo app: Edge Power Scheduler
* install Nexus Runtime and demo app on the cluster
* scale the demo application

# Pre-Requisites

### 1. K8s Cluster
For this scale demo workflow, we will need as K8s cluster with the following following nodes, with the specified taints and labels.

#### 3 Nodes for API Server
```
    labels: { ..., nexus: api }
    taints:
      - {
          "key": "nexus",
          "value": "api",
          "effect": "NoSchedule"
        }
```

#### 3 Nodes for EPS Scheduler App
```
    labels: { ..., eps: scheduler }
    taints:
      - {
          "key": "eps",
          "value": "scheduler",
          "effect": "NoSchedule"
        }
```

#### 3 Nodes for EPS Generator App
```
    labels: { ..., eps: generator }
    taints:
      - {
          "key": "eps",
          "value": "generator",
          "effect": "NoSchedule"
        }
```

#### 8 Nodes for EPS Agent App
```
No special labels / taints for Agent App
```
### 2. Docker

Docker should be installed and running on the machine on which the workflow will be executed.


# Scale Setup Bring up Workflow

## 1. Clone Nexus Repo
```
git clone https://github.com/dilipmighty245/graph-framework-for-microservices.git
cd graph-framework-for-microservices/
export NEXUS_REPO_DIR=${PWD}
```

## 2. Build Nexus Runtime

### Build Nexus Runtime

#### Specify a tag to use for locally built runtime artifacts
```
# The tag can be anything. Here we use a tag called "letsplay"
echo letsplay > TAG
```

#### Build Nexus Runtime
```
make compiler.builder
sudo make runtime.clean; make runtime.build
```

## 3. Build datamodel for demo app: Edge Power Scheduler
```
COMPILER_TAG=letsplay make -C demo/edge-power-scheduler/datamodel datamodel_build
```

## 4. Install the demo

This will deploy Nexus Runtime and Edge Power Scheduler app.

```
make demo.install.kubeconfig
```

### 5. Scale the Edge Power Scheduler App

```
    kubectl scale --replicas=1 statefulset/eps-scheduler
    kubectl scale --replicas=10 statefulset/eps-generator
    kubectl scale --replicas=10 statefulset/eps-edge-agent
```
This will simulate a scaled setup with 2000 Edges.
