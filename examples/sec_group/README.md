# Security Group Example

The Security Group example launches a VPC, a security group, and three rules attached to it.

Note that this example uses the VPC `SecGroup` API through `ucloud_sec_group` and
`ucloud_sec_group_rule`. The `ucloud_security_group` resource is a different object backed by the
`Firewall` API, and the two are not interchangeable.

Each rule is a resource of its own, so changing one is an edit that keeps its rule ID rather than a
delete and a create. Adding a rule means adding a resource; a rule that carries several CIDR blocks
is still a single rule.

To run, configure your UCloud provider as described in https://www.terraform.io/docs/providers/ucloud/index.html

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply`
