# UCloud Provider Examples

This directory contains a set of examples of using various UCloud services with
Terraform. The examples each have their own README containing more details
on what the example does.

To run any example, clone the repository and run `terraform apply` within
the example's own directory.

For example:

```sh
$ git clone https://github.com/terraform-providers/terraform-provider-ucloud
$ cd terraform-provider-ucloud/examples/instance

export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"

$ terraform apply
...
```
