# njalla_email_forward Resource

Njalla email forward for a given domain.

## Example Usage

```hcl
resource njalla_email_forward example {
  domain = "example.com"
  from = "info"
  to = "someone@example.net"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain mail is forwarded from.
* `from` - (Required) Address mail is forwarded from.
* `to` - (Required) Address mail is forwarded to.

~> **Note** Njalla can't edit email forwards, so changing any argument forces
a new forward.

## Attributes Reference

* `id` - `domain:from:to` of the forward.

## Import

```sh
$ terraform import njalla_email_forward.example example.com:info:someone@example.net
```
