package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"strings"
	"time"
)

func resourceUCloudUFSVolumeMountPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUFSVolumeMountPointCreate,
		Read:   resourceUCloudUFSVolumeMountPointRead,
		Delete: resourceUCloudUFSVolumeMountPointDelete,
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateUFSVolumeName,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"mount_point_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudUFSVolumeMountPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ufsconn
	volumeId := d.Get("volume_id").(string)
	vpcId := d.Get("vpc_id").(string)
	subnetId := d.Get("subnet_id").(string)
	name := d.Get("name").(string)

	req := conn.NewAddUFSVolumeMountPointRequest()
	req.VolumeId = ucloud.String(volumeId)
	req.VpcId = ucloud.String(vpcId)
	req.SubnetId = ucloud.String(subnetId)
	req.MountPointName = ucloud.String(name)
	_, err := conn.AddUFSVolumeMountPoint(req)
	if err != nil {
		return fmt.Errorf("error on creating ufs volume mount point, %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", volumeId, vpcId, subnetId))

	return resourceUCloudUFSVolumeMountPointRead(d, meta)
}

func resourceUCloudUFSVolumeMountPointRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	p := strings.Split(d.Id(), ":")
	if len(p) != 3 {
		return fmt.Errorf("illegal ufs volume mount point id, %s", d.Id())
	}
	resourceSet, err := client.describeUFSVolumeMountPointById(p[0], p[1], p[2])
	if err != nil {
		return fmt.Errorf("error on reading ufs volume mount point %q, %s", d.Id(), err)
	}

	d.Set("volume_id", p[0])
	d.Set("name", resourceSet.MountPointName)
	d.Set("vpc_id", resourceSet.VpcId)
	d.Set("subnet_id", resourceSet.SubnetId)
	d.Set("mount_point_ip", resourceSet.MountPointIp)
	d.Set("create_time", timestampToString(resourceSet.CreateTime))

	return nil
}

func resourceUCloudUFSVolumeMountPointDelete(d *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client := meta.(*UCloudClient)
		conn := client.ufsconn

		p := strings.Split(d.Id(), ":")
		if len(p) != 3 {
			return resource.NonRetryableError(fmt.Errorf("illegal ufs volume mount point id, %s", d.Id()))
		}

		req := conn.NewRemoveUFSVolumeMountPointRequest()
		req.VolumeId = ucloud.String(p[0])
		req.VpcId = ucloud.String(p[1])
		req.SubnetId = ucloud.String(p[2])

		if _, err := conn.RemoveUFSVolumeMountPoint(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ufs volume mount point %q, %s", d.Id(), err))
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{statusPending},
			Target:  []string{statusDELETED},
			Refresh: func() (interface{}, string, error) {
				resp, err := client.describeUFSVolumeMountPointById(p[0], p[1], p[2])
				if err != nil {
					if isNotFoundError(err) {
						return resp, statusDELETED, nil
					}
					return nil, statusPending, err
				}
				return resp, statusPending, nil
			},
			Timeout:    2 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 1 * time.Second,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			if _, ok := err.(*resource.TimeoutError); ok {
				return resource.RetryableError(fmt.Errorf("error on waiting for deleting ufs volume mount point %q, %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("error on waiting for deleting ufs volume mount point %q, %s", d.Id(), err))
		}
		return nil
	})
}
