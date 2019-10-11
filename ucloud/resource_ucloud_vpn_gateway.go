package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
)

func resourceUCloudVPNGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPNGatewayCreate,
		Read:   resourceUCloudVPNGatewayRead,
		Update: resourceUCloudVPNGatewayUpdate,
		Delete: resourceUCloudVPNGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"grade": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"standard",
					"enhanced",
				}, false),
			},

			"eip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
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
				Optional: true,
				ForceNew: true,
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

func resourceUCloudVPNGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewCreateVPNGatewayRequest()
	req.EIPId = ucloud.String(d.Get("eip_id").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.Grade = ucloud.String(upperCamelCvt.unconvert(d.Get("grade").(string)))

	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}
	if v, ok := d.GetOk("name"); ok {
		req.VPNGatewayName = ucloud.String(v.(string))
	} else {
		req.VPNGatewayName = ucloud.String(resource.PrefixedUniqueId("tf-vpn-gateway-"))
	}

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
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

	resp, err := conn.CreateVPNGateway(req)
	if err != nil {
		return fmt.Errorf("error on creating vpn gateway, %s", err)
	}

	d.SetId(resp.VPNGatewayId)
	return resourceUCloudVPNGatewayRead(d, meta)
}

func resourceUCloudVPNGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	d.Partial(true)
	if d.HasChange("grade") && !d.IsNewResource() {
		req := conn.NewUpdateVPNGatewayRequest()
		req.Grade = ucloud.String(upperCamelCvt.unconvert(d.Get("grade").(string)))
		req.VPNGatewayId = ucloud.String(d.Id())
		if _, err := conn.UpdateVPNGateway(req); err != nil {
			return fmt.Errorf("error on %s to vpn gateway %q, %s", "UpdateVPNGateway", d.Id(), err)
		}
		d.SetPartial("grade")
	}
	d.Partial(false)

	return resourceUCloudVPNGatewayRead(d, meta)
}
func resourceUCloudVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	vgSet, err := client.describeVPNGatewayById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpn gateway %q, %s", d.Id(), err)
	}

	d.Set("name", vgSet.VPNGatewayName)
	d.Set("remark", vgSet.Remark)
	d.Set("tag", vgSet.Tag)
	d.Set("vpc_id", vgSet.VPCId)
	d.Set("grade", upperCamelCvt.convert(vgSet.Grade))
	d.Set("eip_id", vgSet.EIPId)
	d.Set("charge_type", upperCamelCvt.convert(vgSet.ChargeType))
	d.Set("create_time", timestampToString(vgSet.CreateTime))
	d.Set("expire_time", timestampToString(vgSet.ExpireTime))

	return nil
}

func resourceUCloudVPNGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewDeleteVPNGatewayRequest()
	req.VPNGatewayId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPNGateway(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpn gateway %q, %s", d.Id(), err))
		}

		_, err := client.describeVPNGatewayById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpn gateway when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vpn gateway %q has not been deleted due to unknown error", d.Id()))
	})
}
