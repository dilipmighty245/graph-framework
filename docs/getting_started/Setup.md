# Development setup

Clone the repository using its public HTTPS URL:

```sh
git clone https://github.com/dilipmighty245/graph-framework-for-microservices.git
cd graph-framework-for-microservices
export NEXUS_REPO_DIR="$PWD"
```

Install Go 1.27, Git, Make, and Python 3 for source development. Race tests also
need a platform C toolchain. For generated-model and service tests, install
Docker, Kind, kubectl, jq, yq v4, and goimports and start Docker.

Run the source checks from the repository root:

```sh
make test.framework
```

Run the repeated framework checks, including generation and disposable Kind
clusters, with:

```sh
make test.framework.stability
```

See [testing](../testing.md) for tool prerequisites, coverage, and logs, and the
[framework overview](../../README.md) for its architecture and components.
