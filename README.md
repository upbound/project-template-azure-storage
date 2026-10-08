# project-template-azure-storage

This template can be used to initialize a new project using `provider-azure`. By
default it comes with a namespaced `StorageBucket` XRD (Crossplane v2,
`apiextensions.crossplane.io/v2`) and a matching composition function which
creates an Azure Storage account and container. It also creates the
corresponding unit and e2e tests.

## Usage

To use this template, run the following command:

```shell
up project init -t upbound/project-template-azure-storage --language=kcl <project-name>
```

This template supports the following languages:

- `kcl`
- `go`
- `python`
- `go-templating`
