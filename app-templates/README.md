# Nexus application templates

Application templates scaffold services that consume generated Nexus datamodels. They provide Go source, configuration, Make targets, and container build files for the CLI's application initialization workflow.

## Layout

The `templates` directory contains the project files and templated build configuration. Builder files prepare the development image, while the generated application uses datamodel clients and controller scaffolding.

## Maintaining templates

Validate a rendered project rather than checking only the template source. Initialize a sample application, attach a datamodel, generate the client packages, and compile the application. Build its container if Dockerfiles or platform settings changed.

Review external image defaults and artifact publishing settings before use. Toolchain changes must update both the framework and the project files it generates.

See [CLI workflows](../cli/README.md), [datamodel templates](../datamodel-templates/README.md), and the [framework overview](../README.md).
