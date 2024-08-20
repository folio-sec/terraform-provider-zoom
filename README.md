# Terraform Provider Zoom

- A resource and a data source (`internal/provider/`),
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads)
- [Go](https://golang.org/doc/install)

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make` command:

```shell
make updatespec # to update the spec using the latest version of the Zoom API
make generate # to generate the provider code from openapi spec and the documentation
make build # to build the provider
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

```terraform
terraform {
  required_providers {
    zoom = {
      source  = "folio-sec/zoom"
      version = "~> 0.0.0"
    }
  }
}

provider "zoom" {
  account_id    = var.zoom_account_id
  client_id     = var.zoom_client_id
  client_secret = var.zoom_client_secret
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `dist/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

## Debugging the Provider

You can debug developing provider using following steps:

- Setup your zoom application from https://marketplace.zoom.us/user/build then get the secrets.
- `make local_install`
- Edit `~/.terraformrc` using the output comment
- `cd examples/resources/zoom_phone_autoreceiptionist`
- `TF_LOG_PROVIDER=debug terraform apply`
