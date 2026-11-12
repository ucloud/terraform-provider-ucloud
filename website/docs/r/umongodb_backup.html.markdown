---
subcategory: "UMongoDB"
layout: "ucloud"
page_title: "UCloud: ucloud_umongodb_backup"
description: |-
  Starts a UMongoDB cluster backup.
---

# ucloud_umongodb_backup

Starts one manual backup for an existing UMongoDB cluster. Deleting this resource removes it from Terraform state and does not delete the backup in UCloud.

## Example Usage

```hcl
resource "ucloud_umongodb_backup" "example" {
  availability_zone = ucloud_umongodb_replicaset.example.availability_zone
  cluster_id        = ucloud_umongodb_replicaset.example.id
  backup_name       = "tf-umongodb-backup"
}
```

## Argument Reference

* `availability_zone` - (Required, ForceNew) Zone of the cluster. The backup list does not return this zone, so Read leaves the configured value in place.
* `cluster_id` - (Required, ForceNew) Cluster ID.
* `backup_name` - (Required, ForceNew) Backup name.

## Attributes Reference

* `id` - Backup ID.
* `state` - Backup state, such as `Backup_Success`.
* `backup_size` - Backup size.
* `backup_type` - Backup type.
* `start_time` - Start time as a Unix timestamp.
* `end_time` - End time as a Unix timestamp.

## Import

List backup needs the cluster and zone, so the import ID is `<availability_zone>/<cluster_id>/<backup_id>`.

```
$ terraform import ucloud_umongodb_backup.example cn-bj2-05/umongodb-rs-abc/backup-id
```

`Backup_Stopped` is a stopped backup, not a successful one. Create fails if the backup stops or fails. A sharded snapshot list collapses one batch onto a single backup ID, which can differ from the create response. The provider follows the listed ID with the same backup name. Sharded failed backups are omitted from that list, so a failed sharded backup waits until timeout.
