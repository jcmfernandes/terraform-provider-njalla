# Njalla Provider

The [Njalla][] provider is used to interact with the resources
supported by the [official Njalla API][Njalla API]. The provider needs to be
configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources and data
sources.

## Example Usage

```hcl
# Configure the Njalla provider
terraform {
  required_providers {
    njalla = {
      source = "jcmfernandes/njalla"
    }
  }
}

# Create a TXT Record
resource njalla_record_txt example-txt {
  domain = "example.com"
  name = "example-name"
  ttl = 10800
  content = "example-content"
}
```

## Authentication

The Njalla provider offers a flexible means of providing the API token for
authentication, the following methods are supported, in this order, and
explained below:

* Static credentials
* Environment variables

### Static Credentials

!> **Warning** Hard-coding credentials into any Terraform configuration is
not recommended, and risks secret leakage should this file ever be committed
to a public version control system.

Static credentials can be provided by adding an `api_token` in-line in the
Njalla provider block:

```hcl
provider njalla {
  api_token = "my-api-token"
}
```

### Environment Variables

You can provide the API token via the `NJALLA_API_TOKEN` environment variable.

```hcl
provider njalla {}
```

Usage:

```sh
$ export NJALLA_API_TOKEN='my-api-token'
```

## Argument Reference

* `api_token` - (Optional) This is the Njalla API token. It must be provided,
  but it can also be sourced from the `NJALLA_API_TOKEN` environment variable.

## Limitations

The provider covers DNS records, domains, glue and DNSSEC records, email
forwards, API tokens, servers and VPNs. It doesn't cover renewals, server
start/stop/restart, extra server traffic, `Redirect` and `Static` records,
or the API's user and wallet calls.

!> **Warning** `njalla_domain`, `njalla_server` and `njalla_vpn` charge your
Njalla wallet when created.

[Njalla]: https://njal.la
[Njalla API]: https://njal.la/api/
