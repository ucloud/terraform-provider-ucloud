package ufs

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUFSVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUFSVolumeCreate,
		Read:   resourceUCloudUFSVolumeRead,
		Update: resourceUCloudUFSVolumeUpdate,
		Delete: resourceUCloudUFSVolumeDelete,
		Schema: map[string]*schema.Schema{
			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validateAll(
					validation.IntBetween(100, 100000),
					validateMod(100),
				),
			},
			"storage_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Basic",
					"Advanced",
				}, false),
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"NFSv3",
					"NFSv4",
				}, false),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateUFSVolumeName,
			},
			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"dynamic",
				}, false),
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
			"remark": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expire_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudUFSVolumeCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating ufs volume, %s", err)
	}

	request := client.NewCreateUFSVolumeRequest()
	request.Size = ucloud.Int(data.Get("size").(int))
	request.StorageType = ucloud.String(data.Get("storage_type").(string))
	request.ProtocolType = ucloud.String(data.Get("protocol_type").(string))
	if value, ok := data.GetOk("charge_type"); ok {
		request.ChargeType = ucloud.String(chargeTypeToAPI(value.(string)))
	} else {
		request.ChargeType = ucloud.String("Month")
	}
	if value, ok := data.GetOkExists("duration"); ok {
		request.Quantity = ucloud.Int(value.(int))
	} else {
		request.Quantity = ucloud.Int(1)
	}
	if value, ok := data.GetOk("name"); ok {
		request.VolumeName = ucloud.String(value.(string))
	} else {
		request.VolumeName = ucloud.String(resource.PrefixedUniqueId("tf-ufs-volume-"))
	}
	if value, ok := data.GetOk("tag"); ok {
		request.Tag = ucloud.String(value.(string))
	} else {
		request.Tag = ucloud.String(defaultTag)
	}
	if value, ok := data.GetOk("remark"); ok {
		request.Remark = ucloud.String(value.(string))
	}

	response, err := client.CreateUFSVolume(request)
	if err != nil {
		return fmt.Errorf("error on creating ufs volume, %s", err)
	}
	data.SetId(response.VolumeId)
	return resourceUCloudUFSVolumeRead(data, meta)
}

func resourceUCloudUFSVolumeUpdate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating ufs volume, %s", err)
	}

	data.Partial(true)
	if data.HasChange("size") && !data.IsNewResource() {
		request := client.NewExtendUFSVolumeRequest()
		request.VolumeId = ucloud.String(data.Id())
		request.Size = ucloud.Int(data.Get("size").(int))
		if _, err := client.ExtendUFSVolume(request); err != nil {
			return fmt.Errorf("error on %s to ufs volume %q, %s", "ExtendUFSVolume", data.Id(), err)
		}
		data.SetPartial("size")
	}
	data.Partial(false)
	return resourceUCloudUFSVolumeRead(data, meta)
}

func resourceUCloudUFSVolumeRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading ufs volume, %s", err)
	}

	instance, err := describeUFSVolumeByID(client, data.Id())
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading ufs volume %q, %s", data.Id(), err)
	}

	data.Set("size", instance.Size)
	data.Set("storage_type", instance.StorageType)
	data.Set("protocol_type", instance.ProtocolType)
	data.Set("name", instance.VolumeName)
	data.Set("tag", instance.Tag)
	data.Set("remark", instance.Remark)
	data.Set("create_time", timestampToString(instance.CreateTime))
	data.Set("expire_time", timestampToString(instance.ExpiredTime))
	return nil
}

func resourceUCloudUFSVolumeDelete(data *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client, err := clientFromMeta(meta)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on getting client when deleting ufs volume, %s", err))
		}

		request := client.NewRemoveUFSVolumeRequest()
		request.VolumeId = ucloud.String(data.Id())
		if _, err := client.RemoveUFSVolume(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ufs volume %q, %s", data.Id(), err))
		}
		_, err = describeUFSVolumeByID(client, data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ufs volume when deleting %q, %s", data.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified ufs volume %q has not been deleted due to unknown error", data.Id()))
	})
}
