# Instance Example

The web-server example builds a web server infrastructure, which launches elastic ip and cloud disk to bind UHost instance.

To run, configure your UCloud provider as described in https://www.terraform.io/docs/providers/ucloud/index.html

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply`
