# njalla_api_token Resource

Njalla API token.

## Example Usage

```hcl
resource njalla_api_token acme {
  comment = "certbot"
  from = ["192.0.2.0/24"]
  allowed_domains = ["example.com"]
  acme = true
}
```

## Argument Reference

* `comment` - (Optional) Comment for the token.
* `from` - (Optional) Set of IPv4 or IPv6 addresses or networks allowed to use
  the token.
* `allowed_domains` - (Optional) Set of domains the token is restricted to.
* `allowed_servers` - (Optional) Set of server IDs the token is restricted to.
* `allowed_methods` - (Optional) Set of API methods the token is restricted
  to.
* `allowed_prefixes` - (Optional) Set of DNS record name prefixes the token is
  restricted to.
* `allowed_types` - (Optional) Set of DNS record types the token is restricted
  to.
* `acme` - (Optional) Add the methods, prefixes and types needed for the ACME
  DNS challenge. Default is `false`.

~> **Note** Changing `acme` deletes the token and creates a new one, with a
new key. Other arguments are updated in place.

-> **Note** `acme` fills in `allowed_methods`, `allowed_prefixes` and
`allowed_types`, so removing those from the configuration leaves them
unchanged rather than clearing them.

## Attributes Reference

* `id` - SHA-256 hash of the key.
* `key` - (Sensitive) The API token.

## Import

Tokens are imported by key. Njalla doesn't report whether `acme` was used, so
imported tokens get `acme = false`; setting `acme = true` in the configuration
then replaces the token.

```sh
$ tofu import njalla_api_token.acme <key>
```
