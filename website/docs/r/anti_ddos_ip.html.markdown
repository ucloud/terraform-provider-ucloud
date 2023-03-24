---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_ip"
description: |-
  Provides an Anti-DDoS IP resource.
---

# ucloud_anti_ddos_ip

Provides an Anti-DDoS IP resource.

## Example Usage

```hcl
resource "ucloud_anti_ddos_instance" "foo" {
    area               = "EastChina"
    bandwidth          = 80
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
resource "ucloud_anti_ddos_ip" "foo" {
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    comment = "test-acc-comment"
}
```

## Argument Reference

The following arguments are supported:

* `comment` - (Optional) Comment of the IP.
* `instance_id` - (Required, ForceNew) ID of ucloud_anti_ddos_instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource ucloud_anti_ddos_ip, the format is `<instance_id>/<ip>`.
* `status` - Status of the IP. Possible values are `Pending` and `Success`
* `domain` - Corresponding domain of the IP.

## Import

Anti-DDoS instance allowed domain can be imported using the `<instance_id>/<ip>`, e.g.

```
$ terraform import ucloud_anti_ddos_ip.example usecure_ghp-xxx/10.10.10.10
```
