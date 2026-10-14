# Nexus authorization controller

This controller translates Nexus authorization resources into Kubernetes RBAC resources. It watches roles, bindings, users, and CRDs to maintain access rules for modeled resources.

## Responsibilities

ResourceRole and ResourceRoleBinding resources drive Kubernetes ClusterRole and ClusterRoleBinding creation. Instance role resources describe access at a more specific graph scope. User handling includes certificate creation and storage in user-certificate resources.

Authorization rules depend on datamodel hierarchy and the generated authorization API. Review role expansion and bindings together when changing access behavior.

## Development

The entry point is `cmd/authz-controller`; reconciler implementations are under `controllers`. Tests are under `test` and use Kubernetes resources and certificate fixtures.

Generate the [admin API](../api/README.md) before building. The authorization model lives in `api/authorization` and uses the `authorization.admin.nexus.com` resource group. Configuration owns roles, bindings, users and groups; runtime owns user certificates.

Deployment configuration lives under `pkg/config/deployment` and the `authz` Helm chart. Configure API access, certificate material, image, and namespace for the target environment.

See the [framework overview](../README.md) for the runtime context.
