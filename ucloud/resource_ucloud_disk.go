package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/customdiff"
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

		CustomizeDiff: customdiff.All(
			diffValidateDiskTypeWithZone,
			customdiff.ValidateChange("disk_size", diffValidateDiskSize),
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateDiskName,
			},

			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validateAll(
					validation.IntBetween(1, 4000),
					validateMod(10),
				),
			},

			"disk_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "data_disk",
				ValidateFunc: validation.StringInSlice([]string{"data_disk", "ssd_data_disk", "rssd_data_disk"}, false),
			},

			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "month",
				ValidateFunc: validation.StringInSlice([]string{"year", "month", "dynamic"}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
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

func resourceUCloudDiskCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	req := conn.NewCreateUDiskRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(d.Get("disk_size").(int))
	req.DiskType = ucloud.String(diskTypeCvt.unconvert(d.Get("disk_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-disk-"))
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	resp, err := conn.CreateUDisk(req)
	if err != nil {
		return fmt.Errorf("error on creating disk, %s", err)
	}

	if len(resp.UDiskId) != 1 {
		return fmt.Errorf("error on creating disk, expected exactly one disk, got %v", len(resp.UDiskId))
	}

	d.SetId(resp.UDiskId[0])

	// after create disk, we need to wait it initialized
	stateConf := diskWaitForState(client, d.Id())

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for disk %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudDiskRead(d, meta)
}

func resourceUCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn
	uhostConn := client.uhostconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewRenameUDiskRequest()
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.UDiskId = ucloud.String(d.Id())
		req.UDiskName = ucloud.String(d.Get("name").(string))

		_, err := conn.RenameUDisk(req)
		if err != nil {
			return fmt.Errorf("error on %s to disk %q, %s", "RenameUDisk", d.Id(), err)
		}

		d.SetPartial("name")
	}

	if d.HasChange("disk_size") && !d.IsNewResource() {
		diskSet, err := client.describeDiskById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading disk %q, %s", d.Id(), err)
		}

		if diskSet.UHostId != "" {
			uhostId := diskSet.UHostId
			instance, err := client.describeInstanceById(uhostId)
			if err != nil {
				return fmt.Errorf("error on reading instance %q when updating the size of disk %q, %s", uhostId, d.Id(), err)
			}
			if strings.ToLower(instance.State) != statusStopped {
				stopReq := uhostConn.NewStopUHostInstanceRequest()
				stopReq.UHostId = ucloud.String(uhostId)
				_, err := uhostConn.StopUHostInstance(stopReq)
				if err != nil {
					return fmt.Errorf("error on stopping instance %q when updating the size of disk %q, %s", uhostId, d.Id(), err)
				}

				// after stop instance, we need to wait it stopped
				stateConf := &resource.StateChangeConf{
					Pending:    []string{statusPending},
					Target:     []string{statusStopped},
					Refresh:    instanceStateRefreshFunc(client, uhostId, statusStopped),
					Timeout:    5 * time.Minute,
					Delay:      3 * time.Second,
					MinTimeout: 2 * time.Second,
				}

				if _, err = stateConf.WaitForState(); err != nil {
					return fmt.Errorf("error on waiting for stopping instance %q when updating the size of disk %q, %s", uhostId, d.Id(), err)
				}
			}

			req := uhostConn.NewResizeAttachedDiskRequest()
			req.UHostId = ucloud.String(uhostId)
			req.Zone = ucloud.String(diskSet.Zone)
			req.DiskSpace = ucloud.Int(d.Get("disk_size").(int))
			req.DiskId = ucloud.String(d.Id())
			if _, err := uhostConn.ResizeAttachedDisk(req); err != nil {
				return fmt.Errorf("error on %s to disk %q, %s", "ResizeAttachedDisk", d.Id(), err)
			}
			d.SetPartial("disk_size")

			// after update disk size, we need to wait it completed
			stateConf := diskWaitForState(client, d.Id())

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for %s complete to disk %q, %s", "ResizeAttachedDisk", d.Id(), err)
			}

			instanceAfter, err := client.describeInstanceById(uhostId)
			if err != nil {
				return fmt.Errorf("error on reading instance %q c %q, %s", uhostId, d.Id(), err)
			}

			if strings.ToLower(instanceAfter.State) != statusRunning {
				// after instance update, we need to wait it started
				startReq := uhostConn.NewStartUHostInstanceRequest()
				startReq.UHostId = ucloud.String(uhostId)

				if _, err := uhostConn.StartUHostInstance(startReq); err != nil {
					return fmt.Errorf("error on starting instance %q after updating the size of disk %q, %s", uhostId, d.Id(), err)
				}

				stateConf = &resource.StateChangeConf{
					Pending:    []string{statusPending},
					Target:     []string{statusRunning},
					Refresh:    instanceStateRefreshFunc(client, uhostId, statusRunning),
					Timeout:    d.Timeout(schema.TimeoutUpdate),
					Delay:      3 * time.Second,
					MinTimeout: 2 * time.Second,
				}

				if _, err = stateConf.WaitForState(); err != nil {
					return fmt.Errorf("error on waiting for starting instance %q after updating the size of disk %q, %s", uhostId, d.Id(), err)
				}
			}
		} else {
			req := conn.NewResizeUDiskRequest()
			req.Zone = ucloud.String(d.Get("availability_zone").(string))
			req.UDiskId = ucloud.String(d.Id())
			req.Size = ucloud.Int(d.Get("disk_size").(int))

			_, err := conn.ResizeUDisk(req)
			if err != nil {
				return fmt.Errorf("error on %s to disk %q, %s", "ResizeUDisk", d.Id(), err)
			}
			d.SetPartial("disk_size")

			// after update disk size, we need to wait it completed
			stateConf := diskWaitForState(client, d.Id())

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for %s complete to disk %q, %s", "ResizeUDisk", d.Id(), err)
			}
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
		return fmt.Errorf("error on reading disk %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", diskSet.Zone)
	d.Set("name", diskSet.Name)
	d.Set("tag", diskSet.Tag)
	d.Set("disk_size", diskSet.Size)
	d.Set("charge_type", upperCamelCvt.convert(diskSet.ChargeType))
	d.Set("create_time", timestampToString(diskSet.CreateTime))
	d.Set("expire_time", timestampToString(diskSet.ExpiredTime))
	d.Set("status", diskSet.Status)
	d.Set("disk_type", diskTypeCvt.convert(diskSet.DiskType))

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
			return resource.NonRetryableError(fmt.Errorf("error on deleting disk %q, %s", d.Id(), err))
		}

		_, err := client.describeDiskById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading disk when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified disk %q has not been deleted due to unknown error", d.Id()))
	})
}

func diskWaitForState(client *UCloudClient, diskId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{"available", "inuse"},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			diskSet, err := client.describeDiskById(diskId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			state := strings.ToLower(diskSet.Status)
			if !isStringIn(state, []string{"available", "inuse"}) {
				state = statusPending
			}

			return diskSet, state, nil
		},
	}
}

func diffValidateDiskTypeWithZone(diff *schema.ResourceDiff, v interface{}) error {
	diskType := diff.Get("disk_type").(string)
	zone := diff.Get("availability_zone").(string)

	if diskType == "rssd_data_disk" && zone != "cn-bj2-05" {
		return fmt.Errorf("the disk type about %q only be supported in %q, got %q", "rssd_data_disk", "cn-bj2-05", zone)
	}

	return nil
}

func diffValidateDiskSize(old, new, meta interface{}) error {

	if new.(int) < old.(int) {
		return fmt.Errorf("reduce disk_size is not supported, "+
			"new value %d should be larger than the old value %d", new.(int), old.(int))
	}
	return nil
}
