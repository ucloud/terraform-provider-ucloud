package vpc

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
			"ip_address":  {Type: schema.TypeString, Computed: true},
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudVIPCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating vip, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewAllocateVIPRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.Count = ucloud.Int(1)
	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}
	if value, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(value.(string))
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating vip, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	if !d.IsNewResource() && (d.HasChange("name") || d.HasChange("remark") || d.HasChange("tag")) {
		req := conn.NewUpdateVIPAttributeRequest()
		req.VIPId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))
		req.Remark = ucloud.String(d.Get("remark").(string))
		if value, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(value.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}
		if _, err := conn.UpdateVIPAttribute(req); err != nil {
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading vip, %s", err)
	}
	vip, err := client.describeVIPById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vip %q, %s", d.Id(), err)
	}
	_ = d.Set("name", vip.Name)
	_ = d.Set("tag", vip.Tag)
	_ = d.Set("remark", vip.Remark)
	_ = d.Set("vpc_id", vip.VPCId)
	_ = d.Set("subnet_id", vip.SubnetId)
	_ = d.Set("ip_address", vip.VIP)
	_ = d.Set("create_time", timestampToString(vip.CreateTime))
	return nil
}

func resourceUCloudVIPDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting vip, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewReleaseVIPRequest()
	req.VIPId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.ReleaseVIP(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vip %q, %s", d.Id(), err))
		}
		if _, err := client.describeVIPById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vip when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified vip %q has not been deleted due to unknown error", d.Id()))
	})
}
