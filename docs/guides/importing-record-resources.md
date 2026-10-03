---
page_title: "Importing existing resources"
---

# Importing existing resources

Everything this provider manages can be imported, so objects set up through
the Njalla web interface can be brought under OpenTofu or Terraform. The
examples use `tofu`; `terraform` takes the same arguments.

## Import IDs

| Resource | Import ID |
|----------|-----------|
| `njalla_record_*` | `domain:id` |
| `njalla_dnssec_record` | `domain:id` |
| `njalla_glue_record` | `domain:name` |
| `njalla_email_forward` | `domain:from:to` |
| `njalla_domain` | domain name |
| `njalla_server` | Njalla ID |
| `njalla_vpn` | Njalla ID |
| `njalla_api_token` | token key |

`id` is the object's Njalla ID. Record IDs are listed by the `njalla_records`
data source, server and VPN IDs by `njalla_servers` and `njalla_vpns`. DNSSEC
record IDs are only available from the API's `list-dnssec` method.

Importing a record ID into a resource of another record type isn't detected,
so check the record's `type` first.

## Import blocks

Declare an `import` block next to the resource:

```hcl
import {
  to = njalla_record_txt.example
  id = "example.com:12345"
}

resource njalla_record_txt example {
  domain = "example.com"
  name = "@"
  ttl = 10800
  content = "example-content"
}
```

The next `tofu plan` shows the import, along with any difference between the
configuration and the imported object. `tofu plan
-generate-config-out=generated.tf` writes the resource block for you instead.

## The import command

Alternatively, declare the resource and import it from the command line:

```sh
$ tofu import njalla_record_txt.example example.com:12345
```

This only updates the state. Until the resource block matches the imported
object, plans will change it to match the configuration.

## Caveats

* `njalla_domain`: `years` and `contacts` aren't read back. Importing doesn't
  register or charge anything.
* `njalla_server`: `months` isn't read back.
* `njalla_api_token`: `acme` can't be read back and is imported as `false`.
  Setting `acme = true` in the configuration then replaces the token.

See the [OpenTofu import documentation][OpenTofu import] for details.

[OpenTofu import]: https://opentofu.org/docs/language/import/
