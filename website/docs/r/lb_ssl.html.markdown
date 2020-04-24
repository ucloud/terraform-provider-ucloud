---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_ssl"
sidebar_current: "docs-ucloud-resource-lb-ssl"
description: |-
  Provides a Load Balancer SSL certificate resource.
---

# ucloud_lb_ssl

Provides a Load Balancer SSL certificate resource.

## Example Usage

```hcl
resource "ucloud_lb_ssl" "default" {
  name        = "tf-example-lb-ssl"
  private_key = file("private.key")
  user_cert   = file("user.crt")
  ca_cert     = file("ca.crt")
}
```

## Argument Reference

The following arguments are supported:

* `private_key` - (Required, ForceNew)  The content of the private key about ssl certificate.
* `user_cert` - (Required, ForceNew)  The content of the user certificate about ssl certificate.

- - -

* `name` - (Optional, ForceNew) The name of the LB ssl, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will auto-generate a name beginning with `tf-lb-ssl`.
* `ca_cert` - (Optional, ForceNew) The content of the CA certificate about ssl certificate.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource lb ssl.
* `create_time` - The time of creation for lb ssl, formatted in RFC3339 time string.