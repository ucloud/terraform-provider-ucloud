package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func resourceUCloudDiskAttachment() *schema.Resource {
	return &schema.Resource{
		Create:        resourceUCloudDiskAttachmentCreate,
		Read:          resourceUCloudDiskAttachmentRead,
		Delete:        resourceUCloudDiskAttachmentDelete,
		Update:        schema.Noop,
		SchemaVersion: 1,
		MigrateState:  resourceUCloudDiskAttachmentMigrateState,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"disk_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stop_instance_before_detaching": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceUCloudDiskAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	instanceId := d.Get("instance_id").(string)
	diskId := d.Get("disk_id").(string)

	req := conn.NewAttachUDiskRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.HostId = ucloud.String(instanceId)
	req.UDiskId = ucloud.String(diskId)

	_, err := conn.AttachUDisk(req)
	if err != nil {
		return fmt.Errorf("error on creating disk attachment, %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", diskId, instanceId))

	// after create disk attachment, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{diskStatusAttaching},
		Target:     []string{diskStatusInUse},
		Refresh:    diskAttachmentStateRefreshFunc(client, diskId),
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for disk attachment %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudDiskAttachmentRead(d, meta)
}

func resourceUCloudDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	p := strings.Split(d.Id(), ":")
	resourceSet, err := client.describeDiskResource(p[0], p[1])

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading disk attachment %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", d.Get("availability_zone").(string))
	d.Set("instance_id", resourceSet.UHostId)
	d.Set("disk_id", resourceSet.UDiskId)

	return nil
}

func resourceUCloudDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udiskconn

	p := strings.Split(d.Id(), ":")
	req := conn.NewDetachUDiskRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.UDiskId = ucloud.String(p[0])
	req.HostId = ucloud.String(p[1])

	if _, ok := d.GetOk("stop_instance_before_detaching"); ok {
		err := WaitAndUpdateInstanceState(client, *req.HostId, instanceStatusStopped, false, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return fmt.Errorf("error on stop instance  %q before deleting, %s", *req.HostId, err)
		}
	}

	return resource.Retry(15*time.Minute, func() *resource.RetryError {
		_, err := client.describeDiskResource(p[0], p[1])
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading disk attachment before deleting %q, %s", d.Id(), err))
		}

		if _, err := conn.DetachUDisk(req); err != nil {
			if uErr, ok := err.(uerr.Error); ok && uErr.Code() != 17060 {
				return resource.NonRetryableError(fmt.Errorf("error on deleting disk attachment %q, %s", d.Id(), err))
			}
		}

		// after detach disk, we need to wait it completed
		stateConf := &resource.StateChangeConf{
			Pending:    []string{diskStatusDetaching},
			Target:     []string{diskStatusAvailable},
			Refresh:    diskAttachmentStateRefreshFunc(client, p[0]),
			Timeout:    3 * time.Minute,
			Delay:      2 * time.Second,
			MinTimeout: 1 * time.Second,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			if _, ok := err.(*resource.TimeoutError); ok {
				return resource.RetryableError(fmt.Errorf("error on waiting for deleting disk attachment %q, %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("error on waiting for deleting disk attachment %q, %s", d.Id(), err))
		}

		return nil
	})
}

func diskAttachmentStateRefreshFunc(client *UCloudClient, diskId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		diskSet, err := client.describeDiskById(diskId)
		if err != nil {
			return nil, "", err
		}

		return diskSet, diskSet.Status, nil
	}
}
