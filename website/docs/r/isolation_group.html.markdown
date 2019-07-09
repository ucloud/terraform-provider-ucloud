---
layout: "ucloud"
page_title: "UCloud: ucloud_isolation_group"
sidebar_current: "docs-ucloud-resource-isolation-group"
description: |-
  Provides an Isolation Group resource.
---

# ucloud_isolation_group

Provides an Isolation Group resource. The Isolation Group is a logical group of UHost instance, which ensure that each UHost instance within a group is on a different physical machine. Up to seven UHost instance can be added per isolation group in a single availability_zone.

## Example Usage

```hcl
resource "ucloud_isolation_group" "foo" {
	name  = "tf-acc-isolation-group"
	remark = "test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the isolation group information which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.', ',', '[', ']', ':'. If not specified, terraform will autogenerate a name beginning with `tf-isolation-group`.
* `remark` - (Optional) The remarks of the isolation group. (Default: `""`).

## Import

Isolation Group can be imported using the `id`, e.g.

```
$ terraform import ucloud_isolation_group.example ig-abc123456
```