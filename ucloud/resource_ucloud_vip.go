package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudVIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVIPCreate,
		Read:   resourceUCloudVIPRead,
		Update: resourceUCloudVIPUpdate,
		Delete: resourceUCloudVIPDelete,

		Schema: map[string]*schema.Schema{
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

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip_address": {
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

func resourceUCloudVIPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewAllocateVIPRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.Count = ucloud.Int(1)

	if v, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(v.(string))
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-vip-"))
	}

	resp, err := conn.AllocateVIP(req)
	if err != nil {
		return fmt.Errorf("error on creating vip, %s", err)
	}

	d.SetId(resp.VIPSet[0].VIPId)

	return resourceUCloudVIPRead(d, meta)
}

func resourceUCloudVIPUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	d.Partial(true)

	updateAttribute := false
	req := conn.NewUpdateVIPAttributeRequest()
	req.VIPId = ucloud.String(d.Id())
	if (d.HasChange("name") || d.HasChange("remark") || d.HasChange("tag")) && !d.IsNewResource() {
		updateAttribute = true
	}

	if updateAttribute {
		req.Name = ucloud.String(d.Get("name").(string))
		req.Remark = ucloud.String(d.Get("remark").(string))
		// if tag is empty string, use default tag
		if v, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(v.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}
		_, err := conn.UpdateVIPAttribute(req)
		if err != nil {
			return fmt.Errorf("error on %s to vip %q, %s", "UpdateVIPAttribute", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")
	}

	d.Partial(false)

	return resourceUCloudVIPRead(d, meta)
}

func resourceUCloudVIPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	vip, err := client.describeVIPById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vip %q, %s", d.Id(), err)
	}

	d.Set("name", vip.Name)
	d.Set("tag", vip.Tag)
	d.Set("remark", vip.Remark)
	d.Set("vpc_id", vip.VPCId)
	d.Set("subnet_id", vip.SubnetId)
	d.Set("ip_address", vip.VIP)
	d.Set("create_time", timestampToString(vip.CreateTime))
	return nil
}

func resourceUCloudVIPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewReleaseVIPRequest()
	req.VIPId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.ReleaseVIP(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vip %q, %s", d.Id(), err))
		}
		_, err := client.describeVIPById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vip when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vip %q has not been deleted due to unknown error", d.Id()))
	})
}
