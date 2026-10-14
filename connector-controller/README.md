# Connector controller

Watches NexusEndpoint resources and manages connector deployments, services,
and configuration. The connector itself performs resource replication.

Source is maintained in `connector-controller/` in this repository.
Generate the admin API before building and configure `NAMESPACE` and
`NEXUS_CONNECTOR_VERSION` for the target runtime and connector image.
