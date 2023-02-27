---
layout: "ucloud"
page_title: "Provider: UCloud"
description: |-
  The UCloud provider is used to interact with many resources supported by UCloud. The provider needs to be configured with the proper credentials before it can be used.
---

# UCloud Provider

~> **NOTE:** This guide requires an available UCloud account or sub-account with project to create resources.

The UCloud provider is used to interact with the
resources supported by UCloud. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the terraform provider
terraform {
  required_providers {
    ucloud = {
      source = "ucloud/ucloud"
      version = "~>1.34.1"
    }
  }
}

# Configure the UCloud Provider
provider "ucloud" {
  public_key  = var.ucloud_public_key
  private_key = var.ucloud_private_key
  project_id  = var.ucloud_project_id
  region      = "cn-bj2"
}

# Query default security group
data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

# Query image
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-04"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

# Create web instance 
resource "ucloud_instance" "web" {
  availability_zone = "cn-bj2-04"
  image_id          = data.ucloud_images.default.images[0].id
  instance_type     = "n-basic-2"
  root_password     = "wA1234567"
  name              = "tf-example-instance"
  tag               = "tf-example"
  boot_disk_type    = "cloud_ssd"

  # the default Web Security Group that UCloud recommend to users
  security_group = data.ucloud_security_groups.default.security_groups[0].id

  # create cloud data disk attached to instance
  data_disks {
    size = 20
    type = "cloud_ssd"
  }
  delete_disks_with_instance = true
}
```

## Authentication

The UCloud provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

- Static credentials
- Environment variables

### Static credentials

Static credentials can be provided by adding an `public_key` and `private_key` in-line in the
UCloud provider block:

Usage:

```hcl
terraform {
  required_providers {
    ucloud = {
      source = "ucloud/ucloud"
      version = "~>1.34.1"
    }
  }
}

provider "ucloud" {
  public_key  = "your_public_key"
  private_key = "your_private_key"
  project_id  = "your_project_id"
  region      = "cn-bj2"
}
```

### Environment variables

You can provide your credentials via `UCLOUD_PUBLIC_KEY` and `UCLOUD_PRIVATE_KEY`
environment variables, representing your UCloud public key and private key respectively.
`UCLOUD_REGION` and `UCLOUD_PROJECT_ID` are also used, if applicable:

```hcl
provider "ucloud" {}
```

Usage:

```sh
$ export UCLOUD_PUBLIC_KEY="your_public_key"
$ export UCLOUD_PRIVATE_KEY="your_private_key"
$ export UCLOUD_REGION="cn-bj2"
$ export UCLOUD_PROJECT_ID="org-xxx"

$ terraform plan
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g. `alias` and `version`), the following arguments are supported in the UCloud
 `provider` block:

* `public_key` - (Required) This is the UCloud public key. You may refer to [get public key from console](https://console.ucloud.cn/uapi/apikey). It must be provided, but
  it can also be sourced from the `UCLOUD_PUBLIC_KEY` environment variable.

* `private_key` - (Required) This is the UCloud private key. You may refer to [get private key from console](https://console.ucloud.cn/uapi/apikey). It must be provided, but
  it can also be sourced from the `UCLOUD_PRIVATE_KEY` environment variable.

* `region` - (Required) This is the UCloud region. It must be provided, but
  it can also be sourced from the `UCLOUD_REGION` environment variables.

* `project_id` - (Required) This is the UCloud project id. It must be provided, but
  it can also be sourced from the `UCLOUD_PROJECT_ID` environment variables.

* `max_retries` - (Optional) This is the max retry attempts number. Default max retry attempts number is `0`.

* `insecure` - (Optional) This is a switch to disable/enable https.  (Default: `false`, means enable https). 
 ~> **Note** this argument conflicts with `base_url`.

* `profile` - (Optional) This is the UCloud profile name as set in the shared credentials file, it can also be sourced from the `UCLOUD_PROFILE` environment variables.

* `shared_credentials_file` - (Optional) This is the path to the shared credentials file, it can also be sourced from the `UCLOUD_SHARED_CREDENTIAL_FILE` environment variables. If this is not set and a profile is specified, `~/.ucloud/credential.json` will be used.

* `base_url` - (Optional) This is the base url.(Default: `https://api.ucloud.cn`).
 ~> **Note** this argument conflicts with `insecure`.

## Testing

Credentials must be provided via the `UCLOUD_PUBLIC_KEY`, `UCLOUD_PRIVATE_KEY`, `UCLOUD_PROJECT_ID` environment variables in order to run acceptance tests.
