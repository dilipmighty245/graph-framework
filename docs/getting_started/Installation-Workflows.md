# Nexus: Installation Workflows

## Nexus Runtime Installation

### Installation on K8s Cluster: From Released Artifacts

#### Pre-requisite

1. A K8s cluster
2. KUBECONFIG should be configured in the terminal so K8s cluster is accessible from the terminal

#### Step 1. Download released artifacts needed by Nexus runtime

```
TAG=latest make core.runtime.download
```

NOTE: Modify the value of TAG to use custom release artifacts.

#### Step 2. Upload Nexus runtime artifacts to your Container registry

```
DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make core.runtime.artifacts.push
```
#### Step 3. Install Nexus Runtime

```
DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> NAMESPACE=<k8s-namespace> make core.install.kubeconfig
```

We are DONE !!!

Nexus runtime components now run on the configured K8s namespace.

### Installation on K8s Cluster: From Source Code

#### Pre-requisite

1. A K8s cluster
2. KUBECONFIG should be configured in the terminal so K8s cluster is accessible from the terminal

#### Step 1. Clone the Nexus Repo

```
git clone https://github.com/dilipmighty245/graph-framework-for-microservices.git
```

#### Step 2. Build runtime artifacts from source code

```
TAG=<release-tag> make core.runtime.build
```

#### Step 3. Upload Nexus runtime artifacts to your Container registry

```
DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make core.runtime.artifacts.push
```

#### Step 4. Install Nexus Runtime

```
DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag>  make core.install.kubeconfig
```

We are DONE !!!

Nexus runtime components now run on the configured K8s namespace.

## Datamodel Build + Install

### Build + Install datamodel from pre-built compiler artifact

#### Pre-requisite

1. A K8s cluster
2. KUBECONFIG should be configured in the terminal so K8s cluster is accessible from the terminal
3. Nexus runtime installed on the K8s cluster

#### Step 1: Build datamodel

```
cd < DATAMODEL DIRECTORY >
DATAMODEL_DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make docker.build
```

#### Step 2: Publish datamodel image

```
DATAMODEL_DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make docker.publish
```
#### Step 3: Install datamodel

```
DATAMODEL_DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make dm.install.helm
```

### Build + Install datamodel and compiler from Source code

#### Step 1. Clone the Nexus Repo

```
git clone https://github.com/dilipmighty245/graph-framework-for-microservices.git
cd graph-framework-for-microservices
```

#### Step 2: Build compiler builder

```
make compiler.builder
```

#### Step 3: Build compiler

```
make compiler.build
```

#### Step 4: Build datamodel

```
cd < DATAMODEL DIRECTORY >
DATAMODEL_DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make docker.build
```

#### Step 5: Install datamodel

```
DATAMODEL_DOCKER_REGISTRY=<docker-registry-domain-name> TAG=<release-tag> make dm.install.helm
```
