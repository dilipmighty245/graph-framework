# Nexus admin datamodel

This directory defines the resources used to configure and operate the Nexus runtime. It is a compiler input, with declarations for installed datamodels, gateway routes, authentication, authorization, tenants, users, and replication endpoints.

## Generation

`nexus.yaml` supplies the model group and image repository. The Makefile runs the Nexus compiler and writes generated resources and Go packages to `build/`.

After preparing a compatible compiler image and configuring its registry and tag, run from this directory:

```sh
make datamodel_build
```

Inspect `COMPILER_DOCKER_REGISTRY`, `COMPILER_TAG`, and the Makefile's image settings before running it. Existing defaults are environment-specific.

## Consumers

The gateway, authorization controller, and connector controller import generated resource types and `build/nexus-client`. Build these consumers after generating the admin packages. The authorization model defines roles, bindings, users, and groups under configuration, plus user certificates under runtime.

`make test.framework.e2e` from the repository root generates this model in an isolated snapshot and checks its consumers. See [testing](../docs/testing.md) for prerequisites and coverage.

The module is named `nexus/admin/api` and uses the sibling SDK through a local replacement. Keep the repository layout intact when building. See the [compiler guide](../compiler/README.md) for the generation pipeline.
