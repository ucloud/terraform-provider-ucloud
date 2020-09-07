# Intranet-Cluster Example

The intranet-cluster example builds a intranet cluster infrastructure, which launches a certain amount of UHost instance with the same configuration. It also create vpc and subnet to the Host instances for network isolation.

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
