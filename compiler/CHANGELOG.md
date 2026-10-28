# Compiler release notes

## Unreleased

Generated SDK notification behavior changes:

- Add and delete notifications derive hierarchy relationships from resource
  labels. They no longer wait for parents, poll the API, or skip deletion
  callbacks when a parent is absent. An application that needs the parent must
  handle its absence and retry its own work when appropriate.
- One internal informer handler owns hierarchy updates in event order. Application
  callbacks do not change hierarchy state. Handler goroutines run independently,
  so application callbacks must not assume hierarchy bookkeeping has completed
  when they run. `WaitForCacheSync(ctx)` waits for informer stores, not hierarchy
  handlers or application callbacks.
- Application callbacks retain private object copies. Internal bookkeeping reads
  the cached object's labels and name without making an additional copy.
- Existing callback signatures and LIST/WATCH behavior remain supported.
  The compiler, generated SDK, and runtime client modules now use Kubernetes
  v0.34.3 libraries. Code generation uses v0.34.3 and its `kube_codegen.sh` API;
  controller-runtime consumers use v0.22.3. Regenerate models and rebuild clients
  and GraphQL plugins together. Compiler/generated SDK modules require Go 1.24+.
- Controller manager options use the v0.22 API. Tenant configuration/runtime
  controllers now have distinct names, as required by controller-runtime.
- Enable WatchList through `KUBE_FEATURE_WatchListClient=true` at process startup.
  The library default remains LIST/WATCH; streaming uses client-go's existing
  fallback if unavailable. Both modes are tested with production dependencies.
- Both core charts and Compose select API server v1.34.3 and etcd v3.5.24.
  The legacy chart replaces removed insecure server flags with TLS and a proxy
  that preserves its HTTP Service URLs. Chart certificates are generated per
  installation and reused on Helm upgrades, replacing embedded credentials and
  duplicate bootstrap jobs. Registry mirror overrides and an offline cached-Kind
  image import helper are available. See the runtime guide for credential renewal
  and GitOps configuration. Existing storage migration is a separate operation;
  the compatibility tests use fresh storage.
