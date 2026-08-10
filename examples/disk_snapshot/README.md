# Disk Snapshot Example

The disk snapshot example that creates a ssd cloud disk, takes a snapshot of it, clones a new
cloud disk from that snapshot, and queries the snapshots of the source disk.

The cloned disk leaves `disk_type` unset on purpose: a disk created from `snapshot_id` inherits
the type of its source snapshot.

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
