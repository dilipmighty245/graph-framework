# Contributing

Contributions to Graph Framework for Microservices are managed through this
repository's GitHub issues and pull requests. Start with the [framework
overview](README.md), [compiler guide](compiler/README.md), and
[testing guide](docs/testing.md).

## Proposing a change

For substantial changes, describe the problem, expected behavior, and proposed
scope in an issue before implementing them. Bug reports should include a minimal
reproduction, the affected component, commit or release, operating system, Go
version, and relevant logs. Remove credentials and private data from shared logs.

Documentation corrections and small fixes can be submitted directly as pull
requests. Keep each change focused so reviewers can assess its behavior and tests.

## Development workflow

1. Fork or clone this repository and create a topic branch.
2. Keep sibling modules together: several `go.mod` files use local replacements.
3. Change the source datamodel or generator template before regenerating output.
4. Run the checks appropriate to the affected components.
5. Submit a pull request describing the problem, resulting behavior, and validation.

Use the commands in [docs/testing.md](docs/testing.md):

```sh
make test.framework
make test.framework.compiler
```

The default target runs regression checks that do not need generated admin APIs.
Gateway compilation is a separate check with generation prerequisites. Full
compiler tests currently have documented fixture failures; include the observed
results in the pull request and distinguish existing failures from regressions.
For changes to generation or runtime behavior, also run the relevant Kind checks.

Do not commit executables, caches, credentials, or incidental generated output.
Some checked-in generated files are test fixtures: retain them and review any
intentional regeneration diff. Preserve third-party licenses, copyright notices,
and attribution when editing imported or adapted code.

## Commits and pull requests

Use a short, descriptive subject, such as `fix: preserve parent labels during
replication`. Explain the reason for the change when it is not evident from the
subject. Pull requests should state:

- The concrete problem and the behavior after the change.
- The affected modules, deployment paths, or generated artifacts.
- Checks run and their results, including failures and missing prerequisites.
- Any migration steps required by consumers.

First-party review ownership is recorded in [.github/CODEOWNERS](.github/CODEOWNERS).
Use GitHub issue and pull request discussions to coordinate work; this fork does
not use the former vendor Jira synchronization or vendor CLA-signing links.

## License and attribution

The repository's [LICENSE](LICENSE) and [NOTICE](NOTICE) describe its licensing and
attribution. Submit material you are entitled to contribute and identify any
third-party code or assets included in a change. Preserve the license and notice
files with redistributed source.
