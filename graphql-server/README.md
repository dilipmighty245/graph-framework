# GraphQL server

Watches admin Datamodel resources and loads their generated GraphQL server plugins.
The plugins must be built with compatible Go and dependency versions.

Source is maintained in `graphql-server/` in this repository. Set `KUBECONFIG`
for the target runtime and `PLUGIN_PATH` to the downloaded plugin location.
Use `make help` to inspect the available build targets.
