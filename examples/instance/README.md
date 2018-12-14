# Instance Example

The instance example launches Host instance, the count parameter in variables.tf can let you create specify number Host instances. It also create vpc, subnet, security group for the Host instance.

To run, configure your UCloud provider as described in https://www.terraform.io/docs/providers/ucloud/index.html

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply`
