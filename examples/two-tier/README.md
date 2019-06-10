# Basic Two-Tier UCloud Architecture

This provides a template for running a simple two-tier architecture. The premise is that you have stateless app servers running behind
an ULB serving traffic.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply -var 'count=2'`
