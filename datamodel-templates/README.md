# Nexus datamodel templates

These templates provide a starting layout for a datamodel project and supporting runtime files. The CLI downloads packaged versions of these artifacts when initializing projects.

## Contents

- `nexus/.datamodel.templatedir`: files rendered when creating a model.
- `nexus/helloworld`: example model and packaging files.
- `nexus/runtime-manifests`: runtime files included in the template distribution.
- `nexus/Makefile`: workflow targets shipped with a project.

## Packaging

The top-level Makefile builds template archives and publishes them to an artifact bucket. Inspect `BUCKET` and `TAG` before publishing; the publish target writes to remote storage.

Changes need two checks: the archive must contain the intended files, and a project initialized from that archive must generate and build successfully. Keep generated module paths, sibling dependencies, compiler image settings, and Go versions aligned.

The runtime files here overlap with [nexus-runtime-manifests](../nexus-runtime-manifests/README.md); check both distribution paths when updating deployment behavior.

See the [DSL reference](../compiler/DSL.md) and [compiler guide](../compiler/README.md) to develop a model.
