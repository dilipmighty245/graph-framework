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
  WatchList is still an isolated test experiment; production dependency pins are
  unchanged.
