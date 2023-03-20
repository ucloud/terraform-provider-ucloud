---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_rule"
description: |-
  Provides an Anti-DDoS rule resource.
---

# ucloud_anti_ddos_rule

Provides an Anti-DDoS rule resource.

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
resource "ucloud_anti_ddos_rule" "foo" {
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    ip = "${ucloud_anti_ddos_ip.foo.ip}"
    port = 4321
    real_server_type = "IP"
	real_servers {
      address   = "127.0.0.1"
      port      = 4321
    }
    real_servers {
      address   = "127.0.0.2"
      port      = 4321
    }
	toa_id = 100
	real_server_detection = true
    backup_server = {
        "ip"   = "127.0.0.1"
        "port" = "4321"
    }
}
```

## Argument Reference

The following arguments are supported:

* `comment` - (Optional) Comment of the rule.
* `instance_id` - (Required, ForceNew) ID of ucloud_anti_ddos_instance.
* `ip` -  (Required, ForceNew) IP, where the rule is applied. The value can be got from the IP attribute of ucloud_anti_ddos_ip.
* `port` -  (Optional, ForceNew) Port, where the rule is applied. When port is not set or set zero, the Anti-DDoS instance just forwards traffic in layer-3, otherwise it forwards traffic in layer-4.
* `real_server_type` - (Required, ForceNew) Type of real server address, whose value can be `IP` or `Domain`.
* `real_servers` - (Required) Real server list.
* `toa_id` - (Optional) ID of TOA for getting real client IP. Default is 200.
* `real_server_detection` - (Optional) Whether to detect real server health status. Default is `false`.
* `backup_server` -  (Optional) Backup server, which must be set when `real_server_detection` is `true`.

The `real_servers` object supports the following:
* `address` - (Required) Real server IP.
* `port` - (Optional) Real server port.

The `backup_server` object supports the following:
* `ip` - (Required) Backup server IP.
* `port` - (Optional) Backup server port.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource ucloud_anti_ddos_rule, the format is `<instance_id>/<ip>` for IP protocol rule or  `<instance_id>/<ip>/<port>` for TCP protocol rule.
* `status` - Status of the IP. Possible values are `Pending`, `Success` and `Failed`.

## Import

Anti-DDoS instance allowed domain can be imported using the `<instance_id>/<ip>` or `<instance_id>/<ip>/<port>`, e.g.

```
$ terraform import ucloud_anti_ddos_rule.example usecure_ghp-xxx/10.10.10.10/4321
```
