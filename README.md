# Unofficial Njalla Provider for OpenTofu and Terraform

> [!IMPORTANT]  
> This repository is a **hard fork** of https://github.com/Sighery/terraform-provider-njalla.  

[Njalla][] is a privacy-oriented domain name registration service, with an
[official API][Njalla API].

This repository is an unofficial OpenTofu/Terraform provider for the Njalla
API. It manages DNS records, domains, glue and DNSSEC records, email forwards,
API tokens, servers and VPNs, and reads them through data sources. It talks to
the API through its own small client in [`internal/client`][].

---

## Installing

Releases are published to the [OpenTofu registry][]:

```hcl
terraform {
  required_providers {
    njalla = {
      source = "jcmfernandes/njalla"
    }
  }
}
```

Configure it with a `provider njalla {}` block; see the [documentation][].

## Documentation

The Markdown files in [`docs/`][] document the provider configuration, every
resource and every data source.

---

## Contributing

Resources call the Njalla API through [`internal/client`][]: `Request` for
any method, plus typed helpers for the calls several resources share. Here's
how to add new resources to the provider.

### New resource

Add any new resources inside the `njalla` package. The file name must follow
this format: `resource_{type}`. In the case of our `njalla_record_txt`
resource, the file is then called [`resource_record_txt.go`][]. Data sources
follow `data_source_{type}`. Take a look at any of the existing resources, and
how they're linked in [`provider.go`][].

You'll have to implement the basic `CRUD` operations, and if possible, **do
implement importing as well**.

### Acceptance tests

After adding your new resource (or before), **add acceptance tests**.
Take a look at the
[documentation][Terraform provider acceptance tests documentation] to learn
more. [This Medium article][Terraform provider acceptance tests article]
helped me greatly to understand the structure of acceptance tests and the
whole complex system around them.

Do take a look at existing acceptance tests for the implemented record types,
and copy however much is useful, editing when needed. When copying tests,
**remember to add tests for any new or specific functionality to the new
resource**. For instance: If your new resource has a new field that can only
take certain values, write a new acceptance test for that functionality
specific to that new record.

### Running acceptance tests

These tests **will deploy new infrastructure**. They might fail, and leave
detached/dangling infrastructure, especially during development if your tests
are not yet working properly. It's up to you to clean up afterwards if this is
the case. Any acceptance tests after the development/testing phase is over
should not ever leave dangling resources. Please do test extensively before
making a pull request.

CI doesn't run the tests; run them locally before pushing.

Acceptance tests only run when `TF_ACC` is set, and need two more environment
variables:

* `NJALLA_API_TOKEN`: Njalla API token used to call the API during tests.
* `NJALLA_TESTACC_DOMAIN`: Njalla domain used during the tests.

[mise][] pins Go, OpenTofu and GoReleaser in `mise.toml`. Its `testacc` task
sets `TF_ACC` and runs the tests against OpenTofu:

```bash
export NJALLA_API_TOKEN="api-token-here"
export NJALLA_TESTACC_DOMAIN="testdomain.com"
mise run testacc
```

`mise run test` runs only the unit tests, and `mise run build` builds the
provider.

Two more environment variables opt into tests that are skipped by default:

* `NJALLA_TESTACC_PAID`: Run tests that cost money. They register a random
  `.com` domain (which can't be deleted afterwards), and buy a server and a
  VPN.
* `NJALLA_TESTACC_DNSSEC`: Run tests that publish made-up DNSSEC records for
  `NJALLA_TESTACC_DOMAIN`. A record left behind by a failed run breaks
  resolution of the domain.

### Releasing

There's a [Github Action set up to handle releases][Action Release] on tag
pushes. This action then makes use of [GoReleaser][] to cross-compile to
different platforms. GoReleaser is also used to create a checksums file, and
create a new draft GitHub release.

Before publishing the draft, sign the checksums file with the provider's GPG
key and upload the signature to the release. The OpenTofu registry picks up
published releases automatically, and rejects any without a valid signature.

The provider and its GPG key are registered with the OpenTofu registry once,
through its [submission forms][OpenTofu registry submission].

[Njalla]: https://njal.la
[Njalla API]: https://njal.la/api/
[`internal/client`]: internal/client/
[OpenTofu registry]: https://search.opentofu.org/
[documentation]: docs/index.md
[`docs/`]: docs/
[`resource_record_txt.go`]: njalla/resource_record_txt.go
[`provider.go`]: njalla/provider.go
[Terraform provider acceptance tests documentation]: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests
[Terraform provider acceptance tests article]: https://medium.com/spaceapetech/creating-a-terraform-provider-part-2-1346f89f082c
[Action Release]: .github/workflows/release.yml
[GoReleaser]: https://goreleaser.com/
[mise]: https://mise.jdx.dev
[OpenTofu registry submission]: https://github.com/opentofu/registry#readme
