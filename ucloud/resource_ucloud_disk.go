package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudDiskCreate,
		Read:   resourceUCloudDiskRead,
		Update: resourceUCloudDiskUpdate,
		Delete: resourceUCloudDiskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resource.PrefixedUniqueId("tf-disk-"),
				ValidateFunc: validateDiskName,
			},

			"disk_size": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 4000),
			},

			"disk_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "data_disk",
				ValidateFunc: validation.StringInSlice([]string{"data_disk", "ssd_data_disk"}, false),
			},

			"charge_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "month",
				ValidateFunc: validation.StringInSlice([]string{"year", "month", "dynamic"}, false),
			},

			"duration": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validateDuration,
			},

			"tag": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateTag,
			},

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudDiskCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	req := conn.NewCreateUDiskRequest()
	req.Name = ucloud.String(d.Get("name").(string))
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(d.Get("disk_size").(int))
	req.DiskType = ucloud.String(upperCamelCvt.unconvert(d.Get("disk_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Quantity = ucloud.Int(d.Get("duration").(int))

	if val, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(val.(string))
	}

	resp, err := conn.CreateUDisk(req)
	if err != nil {
		return fmt.Errorf("error on creating disk, %s", err)
	}

	if len(resp.UDiskId) > 0 {
		d.SetId(resp.UDiskId[0])
	}

	// after create disk, we need to wait it initialized
	stateConf := diskWaitForState(client, d.Id())

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for disk %s complete creating, %s", d.Id(), err)
	}

	return resourceUCloudDiskRead(d, meta)
}

func resourceUCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewRenameUDiskRequest()
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.UDiskId = ucloud.String(d.Id())
		req.UDiskName = ucloud.String(d.Get("name").(string))

		_, err := conn.RenameUDisk(req)
		if err != nil {
			return fmt.Errorf("error on %s to disk %s, %s", "RenameUDisk", d.Id(), err)
		}

		d.SetPartial("name")
	}

	if d.HasChange("disk_size") && !d.IsNewResource() {
		req := conn.NewResizeUDiskRequest()
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.UDiskId = ucloud.String(d.Id())
		req.Size = ucloud.Int(d.Get("disk_size").(int))

		_, err := conn.ResizeUDisk(req)
		if err != nil {
			return fmt.Errorf("error on %s to disk %s, %s", "ResizeUDisk", d.Id(), err)
		}

		d.SetPartial("disk_size")

		// after update disk size, we need to wait it completed
		stateConf := diskWaitForState(client, d.Id())

		if _, err = stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for %s complete to disk %s, %s", "ResizeUDisk", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudDiskRead(d, meta)
}

func resourceUCloudDiskRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	diskSet, err := client.describeDiskById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading disk %s, %s", d.Id(), err)
	}

	d.Set("availability_zone", diskSet.Zone)
	d.Set("name", diskSet.Name)
	d.Set("tag", diskSet.Tag)
	d.Set("disk_size", diskSet.Size)
	d.Set("charge_type", upperCamelCvt.convert(diskSet.ChargeType))
	d.Set("create_time", timestampToString(diskSet.CreateTime))
	d.Set("expire_time", timestampToString(diskSet.ExpiredTime))
	d.Set("status", diskSet.Status)

	return nil
}

func resourceUCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	req := conn.NewDeleteUDiskRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.UDiskId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteUDisk(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting disk %s, %s", d.Id(), err))
		}

		_, err := client.describeDiskById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading disk when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified disk %s has not been deleted due to unknown error", d.Id()))
	})
}

func diskWaitForState(client *UCloudClient, diskId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{"available"},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {
			diskSet, err := client.describeDiskById(diskId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			state := strings.ToLower(diskSet.Status)
			if state != "available" {
				state = statusPending
			}

			return diskSet, state, nil
		},
	}
}
