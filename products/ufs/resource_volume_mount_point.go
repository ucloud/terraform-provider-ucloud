package ufs

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
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

func resourceUCloudUFSVolumeMountPointCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating ufs volume mount point, %s", err)
	}

	volumeID := data.Get("volume_id").(string)
	vpcID := data.Get("vpc_id").(string)
	subnetID := data.Get("subnet_id").(string)
	request := client.NewAddUFSVolumeMountPointRequest()
	request.VolumeId = ucloud.String(volumeID)
	request.VpcId = ucloud.String(vpcID)
	request.SubnetId = ucloud.String(subnetID)
	request.MountPointName = ucloud.String(data.Get("name").(string))
	if _, err := client.AddUFSVolumeMountPoint(request); err != nil {
		return fmt.Errorf("error on creating ufs volume mount point, %s", err)
	}

	data.SetId(fmt.Sprintf("%s:%s:%s", volumeID, vpcID, subnetID))
	return resourceUCloudUFSVolumeMountPointRead(data, meta)
}

func resourceUCloudUFSVolumeMountPointRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading ufs volume mount point, %s", err)
	}

	parts := strings.Split(data.Id(), ":")
	if len(parts) != 3 {
		return fmt.Errorf("illegal ufs volume mount point id, %s", data.Id())
	}
	resourceSet, err := describeUFSVolumeMountPointByID(client, parts[0], parts[1], parts[2])
	if err != nil {
		return fmt.Errorf("error on reading ufs volume mount point %q, %s", data.Id(), err)
	}

	data.Set("volume_id", parts[0])
	data.Set("name", resourceSet.MountPointName)
	data.Set("vpc_id", resourceSet.VpcId)
	data.Set("subnet_id", resourceSet.SubnetId)
	data.Set("mount_point_ip", resourceSet.MountPointIp)
	data.Set("create_time", timestampToString(resourceSet.CreateTime))
	return nil
}

func resourceUCloudUFSVolumeMountPointDelete(data *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client, err := clientFromMeta(meta)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on getting client when deleting ufs volume mount point, %s", err))
		}

		parts := strings.Split(data.Id(), ":")
		if len(parts) != 3 {
			return resource.NonRetryableError(fmt.Errorf("illegal ufs volume mount point id, %s", data.Id()))
		}
		request := client.NewRemoveUFSVolumeMountPointRequest()
		request.VolumeId = ucloud.String(parts[0])
		request.VpcId = ucloud.String(parts[1])
		request.SubnetId = ucloud.String(parts[2])
		if _, err := client.RemoveUFSVolumeMountPoint(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ufs volume mount point %q, %s", data.Id(), err))
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{statusPending},
			Target:  []string{statusDELETED},
			Refresh: func() (interface{}, string, error) {
				response, err := describeUFSVolumeMountPointByID(client, parts[0], parts[1], parts[2])
				if err != nil {
					if isNotFoundError(err) {
						return response, statusDELETED, nil
					}
					return nil, statusPending, err
				}
				return response, statusPending, nil
			},
			Timeout:    2 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 1 * time.Second,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			if _, ok := err.(*resource.TimeoutError); ok {
				return resource.RetryableError(fmt.Errorf("error on waiting for deleting ufs volume mount point %q, %s", data.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("error on waiting for deleting ufs volume mount point %q, %s", data.Id(), err))
		}
		return nil
	})
}
