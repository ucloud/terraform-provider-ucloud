---
layout: "ucloud"
page_title: "UCloud: ucloud_instance_types"
sidebar_current: "docs-ucloud-datasource-instance_types"
description: |-
  Provides build a list of instance types by cpu and memory.
---

# ucloud_instance_types

This data source providers build a list of instance types by cpu and memory.

## Example Usage

```hcl
data "ucloud_instance_types" "example" {
    cpu = 1
    memory = 4
}

output "normal" {
    value = "${data.ucloud_instance_types.example.instance_types.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `cpu` - (Required) The number of cores of virtual CPU, measured in "core", range from 1 to 32.
* `memory` - (Required) The size of memory, measured in MB, range from 1 to 128.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instance_types` - instance_types is a nested type. If the ratio of CPU and memory is normal, then the instance_types contains two elements , one is Customized, another is standard, otherwise, the instance_types contains one element Customized. instance_types documented below.

The attribute (`instance_types`) support the following:

* `id` - The time of creation for EIP.
