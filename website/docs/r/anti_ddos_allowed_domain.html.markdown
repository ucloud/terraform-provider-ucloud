---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_allowed_domain"
description: |-
  Provides an Anti-DDoS instance allowed domain resource.
---

# ucloud_anti_ddos_allowed_domain

Provides an Anti-DDoS instance allowed domain resource.

## Example Usage

```hcl
resource "ucloud_anti_ddos_instance" "foo" {
    area               = "EastChina"
    bandwidth          = 50
    base_defence_value = 30
    data_center        = "Zaozhuang"
    max_defence_value  = 30
    name               = "tf-acc-anti-ddos-instance-basic"
}
resource "ucloud_anti_ddos_allowed_domain" "foo" {
    domain      = "ucloud.cn"
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    comment = "test-acc-comment"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required, ForceNew) For domain like `api.ucloud.cn` the value should be `ucloud.cn`.
* `comment` - (Optional) Comment of the domain.
* `instance_id` - (Required, ForceNew) ID of ucloud_anti_ddos_instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource ucloud_anti_ddos_allowed_domain, the format is `${instance_id}/${domain}`.
* `status` -  The status of ucloud_anti_ddos_instance. Possible values are `Adding`, `Success`, `Deleting`, `Failure` and `Deleted`.

## Import

Anti-DDoS instance allowed domain can be imported using the `${instance_id}/${domain}`, e.g.

```
$ terraform import ucloud_anti_ddos_allowed_domain.example usecure_ghp-xxx/ucloud.cn
```
