---
subcategory: "Cube"
layout: "ucloud"
page_title: "UCloud: ucloud_cube_pod"
description: |-
  Provides a CubePod resource.
---

# ucloud_cube_pod

Provides a CubePod resource.

## Example Usage

```hcl
data "ucloud_zones" "default" {
}
data "ucloud_vpcs" "default" {
  name_regex = "DefaultVPC"
}
data "ucloud_subnets" "default" {
  vpc_id = data.ucloud_vpcs.default.vpcs.0.id
}
resource "ucloud_cube_pod" "foo" {
	name  	 	  = "tf-acc-cube-pod-basic"
	tag           = "tf-acc"
    vpc_id        = data.ucloud_vpcs.default.vpcs.0.id
    subnet_id     = data.ucloud_subnets.default.subnets.0.id
    pod           = file("cube_pod.yml")
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where Cube Pod is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `vpc_id` - (Optional, ForceNew) The ID of VPC linked to the Cube Pod. If not defined `vpc_id`, the Cube Pod will use the default VPC in the current region.
* `subnet_id` - (Optional, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the Cube Pod will use the default subnet in the current region.
* `pod` - (Required) The pod yaml of the Cube Pod.

- - -

* `charge_type` - (Optional, ForceNew) The charge type of Cube Pod, possible values are: `year`, `month` and `postpay` as pay after use. (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the Cube Pod (Default: `1`). The value is `0` when pay by month and the Cube Pod will be valid till the last day of that month. It is not required when `postpay` (pay by hour).
* `name` - (Optional) The name of Cube Pod, expected value to be 6 - 63 characters and only support english, numbers, '-', '_', and can not prefix with '-'. If not specified, terraform will auto-generate a name beginning with `tf-cube-pod`.
* `tag` - (Optional, ForceNew) A tag assigned to CubePod, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource Cube Pod.
* `create_time` - The time of creation of Cube Pod, formatted in RFC3339 time string.
* `expire_time` - The expiration time of Cube Pod, formatted in RFC3339 time string.
* `status` - The Cube Pod current status. 
* `pod_ip` - The private ip of pod.