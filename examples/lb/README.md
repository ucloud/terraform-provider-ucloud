# Load Balance Example

The lb example launches load balance and  attach Host instances. It also create rule for the load balance.

To run, configure your UCloud provider as described in https://www.terraform.io/docs/providers/ucloud/index.html

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply`