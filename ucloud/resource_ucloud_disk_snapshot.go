package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudDiskSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudDiskSnapshotCreate,
		Read:   resourceUCloudDiskSnapshotRead,
		Delete: resourceUCloudDiskSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"disk_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateDiskName,
			},

			// the remote api rejects the Comment parameter of CreateUDiskSnapshot with
			// "Params [Comment] not available", so it is exported as a read only attribute
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"source_disk_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_disk_available": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudDiskSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	req := conn.NewCreateUDiskSnapshotRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.UDiskId = ucloud.String(d.Get("disk_id").(string))

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-disk-snapshot-"))
	}

	resp, err := conn.CreateUDiskSnapshot(req)
	if err != nil {
		return fmt.Errorf("error on creating disk snapshot, %s", err)
	}

	if len(resp.SnapshotId) != 1 {
		return fmt.Errorf("error on creating disk snapshot, expected exactly one snapshot, got %v", len(resp.SnapshotId))
	}

	d.SetId(resp.SnapshotId[0])

	// after create disk snapshot, we need to wait it completed
	stateConf := diskSnapshotWaitForState(client, d.Id())

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for disk snapshot %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudDiskSnapshotRead(d, meta)
}

func resourceUCloudDiskSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	snapshotSet, err := client.describeSnapshotById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading disk snapshot %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", snapshotSet.Zone)
	d.Set("disk_id", snapshotSet.UDiskId)
	d.Set("name", snapshotSet.Name)
	d.Set("comment", snapshotSet.Comment)
	d.Set("disk_type", snapshotDiskTypeCvt.convert(snapshotSet.DiskType))
	d.Set("size", snapshotSet.Size)
	d.Set("source_disk_name", snapshotSet.UDiskName)
	d.Set("is_disk_available", snapshotSet.IsUDiskAvailable)
	d.Set("create_time", timestampToString(snapshotSet.CreateTime))
	d.Set("status", snapshotSet.Status)

	return nil
}

func resourceUCloudDiskSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	req := conn.NewDeleteUDiskSnapshotRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.SnapshotId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteUDiskSnapshot(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting disk snapshot %q, %s", d.Id(), err))
		}

		_, err := client.describeSnapshotById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading disk snapshot when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified disk snapshot %q has not been deleted due to unknown error", d.Id()))
	})
}

func diskSnapshotWaitForState(client *UCloudClient, snapshotId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{snapshotStatusNormal},
		Timeout:    20 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh: func() (interface{}, string, error) {
			snapshotSet, err := client.describeSnapshotById(snapshotId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			if snapshotSet.Status == snapshotStatusFailed {
				return nil, "", fmt.Errorf("the specified disk snapshot %q is in %q status", snapshotId, snapshotStatusFailed)
			}

			state := snapshotSet.Status
			if state != snapshotStatusNormal {
				state = statusPending
			}

			return snapshotSet, state, nil
		},
	}
}
