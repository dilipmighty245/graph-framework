# Nexus calibration tooling

The calibration tool generates REST and GraphQL workloads and collects trace data for performance analysis. It is development and measurement tooling rather than a required runtime service.

## Contents

`rest` and `graphqlcalls` issue requests. `workmanager` coordinates workloads, and `traceparser` processes trace data. Configuration and sample request data are in `conf.yaml` and `rest_data.yaml`.

The `deployment` directory contains example observability and workload deployments. The scripts generate test runs and collect resource counts. `etcd-snapshots` documents a snapshot restore workflow.

## Build

From this directory:

```sh
go build -o bin/calib .
```

Before running a workload, configure the target APIs, credentials, request data, and observability endpoints. The examples assume a particular database and tracing setup; provision and adapt those services first.

Measure against an isolated environment with a recorded datamodel, dataset size, component versions, resource limits, and workload profile so results can be compared.

See the [framework overview](../README.md) for the components under measurement.
