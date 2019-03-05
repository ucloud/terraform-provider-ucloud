# Load Balancer Example

The lb example launches load balancer and  attach Host instances. It also create rule for the load balancer and bind SSL certificate to the load balancer.

To run, configure your UCloud provider as described in https://www.terraform.io/docs/providers/ucloud/index.html

## Setup Environment

```sh
export UCLOUD_PUBLIC_KEY="your public key"
export UCLOUD_PRIVATE_KEY="your private key"
export UCLOUD_PROJECT_ID="your project id"
```

## Running the example

run `terraform apply`