# Example graph to parse

```
                  Root
                   |
                   |
                  Config
                 /     \
                /       \
    softlink  /         \
 Dns------->Gns         AccessControlPolicy
             |                  |
             |                  |
         SvcGroup ----------> ACPConfig
                   softlink

```

This directory is the source model used by compiler unit tests and generated-SDK
integration tests. Keep its Go files, GraphQL queries, `nexus.yaml`, module files,
and Makefile in version control. The `ignored/` and `vendor/` fixtures exercise
the parser's directory filtering.

`build/` contains generated artifacts and is ignored by Git. From the compiler
directory, `make generate_example` generates the example under
`example/output/generated/`; from the repository root, `make test.framework.e2e`
generates it in a temporary checkout and tests it against a fresh Kind cluster.

TODO:
Extend examples with:
1. Link types
- Link1 CustomLinkType `nexus:"link"`
- NamedLink1 CustomLinkType `nexus:"links"`
2. Status type
- status CustomStatusType `nexus:"status"`
3. Custom comments
- rest-api
- version
- validation-endpoints
