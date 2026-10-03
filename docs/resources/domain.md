# njalla_domain Resource

Njalla domain. Creating it registers the domain, and manages its settings
afterwards.

!> **Warning** Creating this resource registers the domain, which charges your
Njalla wallet. Njalla can't delete domains: destroying the resource only
removes it from the Terraform state, and the domain stays in your account
until it expires.

## Example Usage

```hcl
resource njalla_domain example {
  name = "example.com"
  years = 1
  lock = true
  nameservers = ["ns1.example.net", "ns2.example.net"]
}
```

## Argument Reference

* `name` - (Required) Domain name to register.
* `years` - (Optional) Years to register the domain for. Default is `1`. Only
  used when registering; changes afterwards are ignored.
* `mailforwarding` - (Optional) Whether email forwarding is enabled. Left
  unmanaged if unset.
* `dnssec` - (Optional) Whether DNSSEC is enabled. Left unmanaged if unset.
* `lock` - (Optional) Whether the domain is locked against transfers. Left
  unmanaged if unset.
* `contacts` - (Optional) Map of custom WHOIS contact IDs. Only sent to
  Njalla, never read back.
* `nameservers` - (Optional) Custom nameservers. Unset or empty uses Njalla's
  nameservers.

~> **Note** Changing `name` registers a new domain. The previous one stays in
your account.

-> **Note** Renewals aren't managed. Renew through the Njalla web interface or
API.

## Attributes Reference

* `id` - The domain name.
* `status` - Status of the domain.
* `expiry` - Expiry date of the domain.
* `max_nameservers` - Maximum number of custom nameservers.

## Timeouts

* `create` - (Default `30m`) How long to wait for the registration to finish.

## Import

Already owned domains can be imported by name, without registering them:

```sh
$ terraform import njalla_domain.example example.com
```
