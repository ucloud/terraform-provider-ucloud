package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
)

func resourceUCloudUFS() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUFSCreate,
		Read:   resourceUCloudUFSRead,
		Update: resourceUCloudUFSUpdate,
		Delete: resourceUCloudUFSDelete,
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
				ValidateFunc: validateName,
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

func resourceUCloudUFSCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ufsconn
	req := conn.NewCreateUFSVolumeRequest()
	req.Size = ucloud.Int(d.Get("size").(int))
	req.StorageType = ucloud.String(d.Get("storage_type").(string))
	req.ProtocolType = ucloud.String(d.Get("protocol_type").(string))
	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("name"); ok {
		req.VolumeName = ucloud.String(v.(string))
	} else {
		req.VolumeName = ucloud.String(resource.PrefixedUniqueId("tf-ufs-"))
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	if v, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(v.(string))
	}
	resp, err := conn.CreateUFSVolume(req)
	if err != nil {
		return fmt.Errorf("error on creating ufs, %s", err)
	}

	d.SetId(resp.VolumeId)

	return resourceUCloudUFSRead(d, meta)
}

func resourceUCloudUFSUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ufsconn

	d.Partial(true)

	if d.HasChange("size") && !d.IsNewResource() {
		reqBand := conn.NewExtendUFSVolumeRequest()
		reqBand.VolumeId = ucloud.String(d.Id())
		reqBand.Size = ucloud.Int(d.Get("size").(int))

		_, err := conn.ExtendUFSVolume(reqBand)
		if err != nil {
			return fmt.Errorf("error on %s to ufs %q, %s", "ExtendUFSVolume", d.Id(), err)
		}

		d.SetPartial("size")
	}

	d.Partial(false)

	return nil
}

func resourceUCloudUFSRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instance, err := client.describeUFSById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading ufs %q, %s", d.Id(), err)
	}

	d.Set("size", instance.Size)
	d.Set("storage_type", instance.StorageType)
	d.Set("protocol_type", instance.ProtocolType)
	d.Set("name", instance.VolumeName)
	d.Set("tag", instance.Tag)
	d.Set("remark", instance.Remark)
	d.Set("create_time", timestampToString(instance.CreateTime))
	d.Set("expire_time", timestampToString(instance.ExpiredTime))

	return nil
}

func resourceUCloudUFSDelete(d *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client := meta.(*UCloudClient)
		conn := client.ufsconn

		req := conn.NewRemoveUFSVolumeRequest()
		req.VolumeId = ucloud.String(d.Id())
		if _, err := conn.RemoveUFSVolume(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ufs %q, %s", d.Id(), err))
		}

		_, err := client.describeUFSById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ufs when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified ufs %q has not been deleted due to unknown error", d.Id()))
	})
}
