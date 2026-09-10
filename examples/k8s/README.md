# UK8S Example

The example creates a cluster with an initial Worker, a node group, and two
additional Workers associated with the group.

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

Use a region code such as `sg` for `region` and a zone such as `sg-02` for `zone`.
Choose image IDs and a Kubernetes version available in that region before
applying. The sample uses the O2 Intel family and RSSD disks.

```sh
terraform init
terraform plan
terraform apply
```

For existing cluster configurations, `master.instance_type` remains supported
but is deprecated. Keep that field when upgrading the provider to avoid
replacing the cluster merely to switch specification syntax. Use the independent
Master fields shown in `main.tf` for new clusters.

To import an existing node group, use its cluster ID and node group ID:

```sh
terraform import ucloud_uk8s_node_group.workers uk8s-example/nodegroup-example
```
