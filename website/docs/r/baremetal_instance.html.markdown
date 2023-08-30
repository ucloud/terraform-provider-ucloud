---
subcategory: "Bare Metail (UPhost)"
layout: "ucloud"
page_title: "UCloud: ucloud_baremetal_instance"
sidebar_current: "docs-ucloud-resource-baremetal-instance"
description: |-
  Provides a UCloud Bare Metal instance resource.

---

# ucloud_baremetal_instance

The `ucloud_baremetal_instance` resource provides a UCloud Bare Metal instance. This can be used to create, modify, and delete Bare Metal instances.

## Example Usage
```hcl
resource "ucloud_baremetal_instance" "example" {
availability_zone = "cn-bj2-02"
image_id = "uimage-abc12345"
password = "your-password"
instance_type = "n-metal-2"
name = "baremetal_example"
}
```

## Argument Reference

The following arguments are supported:

- `availability_zone` - (Required) The availability zone where the instance is created.
- `image_id` - (Required) The ID of the image used to launch the instance.
- `password` - (Required) Password to an instance. Must contain at least 2 types of the following: Lower case letters, upper case letters, digits and special characters. Special characters such as `(`, `)`, `<`, `>`, `,`, `.` , `;`, `:`, `?`, `/` are not allowed. The length of password must be 12~36.
- `instance_type` - (Required) The type of instance to start.
- `name` - (Optional) Name of the instance, which cannot be longer than 63 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the instance.
- `private_ip` - The private IP of the instance.
- `public_ip` - The public IP of the instance.
- `status` - The status of the instance.