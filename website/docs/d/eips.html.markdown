---
layout: "ucloud"
page_title: "UCloud: ucloud_eips"
sidebar_current: "docs-ucloud-datasource-eips"
description: |-
  Provides a list of EIP resources in the current region.
---

# ucloud_eips

This data source provides a list of EIP resources (Elastic IP address) according to their EIP ID.

## Example Usage

```hcl
data "ucloud_eips" "example" {}

output "first" {
    value = "${data.ucloud_eips.example.eips.0.ip_set.0.ip}"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) The ID of Elastic IP, all the EIPs belong to this region will be retrieved if the ID is "".
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `eips` - eips is a nested type. eips documented below.
* `total_count` - Total number of Elastic IP that satisfy the condition.

The attribute (`eips`) support the following:

* `bandwidth` - Maximum bandwidth to the elastic public network, measured in Mbps, this attribute will be displayed as shared bandwidth if "BandwidthType=1", otherwise it will be displayed as the bandwith to EIP if "BandwidthType=0".
* `ip_set` - ip_set is a nested type. ip_set documented below.
* `create_time` - The time of creation for Elastic IP.
* `expire_time` - The expiration time for Elastic IP, formatted by Unix Timestamp.
* `internet_charge_mode` - Elastic IP charge mode. Possible values are: "Traffic" as pay by traffic, "Bandwidth" as pay by bandwidth, "ShareBandwidth" as pay by shared bandwidth.
* `internet_charge_type` - Elastic IP Charge type. Possible values are: "Year" as pay by year, "Month" as pay by month, "Dynamic" as pay by hour.
* `name` - The name of Elastic IP.
* `remark` - The remarks of Elastic IP.
* `status` - Elastic IP status. Possible values are: "used" as in use, "free" as available and "freeze" as associating.
* `tag` - A mapping of tags to assign to the Elastic IP.

The attribute (`ip_set`) support the following:

* `internet_type` - Elastic IP routes. Possible values are: "International" as internaltional IP and "Bgp" as BGP IP.
* `ip` - Elastic IP address