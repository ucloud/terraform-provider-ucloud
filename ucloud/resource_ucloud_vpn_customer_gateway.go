package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
)

func resourceUCloudVPNCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPNCustomerGatewayCreate,
		Read:   resourceUCloudVPNCustomerGatewayRead,
		Delete: resourceUCloudVPNCustomerGatewayDelete,

		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.SingleIP(),
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
				Computed: true,
				ForceNew: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudVPNCustomerGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewCreateRemoteVPNGatewayRequest()
	req.RemoteVPNGatewayAddr = ucloud.String(d.Get("ip_address").(string))
	if v, ok := d.GetOk("name"); ok {
		req.RemoteVPNGatewayName = ucloud.String(v.(string))
	} else {
		req.RemoteVPNGatewayName = ucloud.String(resource.PrefixedUniqueId("tf-vpn-costomer-gateway-"))
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

	resp, err := conn.CreateRemoteVPNGateway(req)
	if err != nil {
		return fmt.Errorf("error on creating vpn costomer gateway, %s", err)
	}

	d.SetId(resp.RemoteVPNGatewayId)
	return resourceUCloudVPNCustomerGatewayRead(d, meta)
}

func resourceUCloudVPNCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	vgSet, err := client.describeVPNCustomerGatewayById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpn costomer gateway %q, %s", d.Id(), err)
	}

	d.Set("name", vgSet.RemoteVPNGatewayName)
	d.Set("remark", vgSet.Remark)
	d.Set("tag", vgSet.Tag)
	d.Set("ip_address", vgSet.RemoteVPNGatewayAddr)
	d.Set("create_time", timestampToString(vgSet.CreateTime))

	return nil
}

func resourceUCloudVPNCustomerGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewDeleteRemoteVPNGatewayRequest()
	req.RemoteVPNGatewayId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteRemoteVPNGateway(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpn costomer gateway %q, %s", d.Id(), err))
		}

		_, err := client.describeVPNCustomerGatewayById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpn costomer gateway when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vpn costomer gateway %q has not been deleted due to unknown error", d.Id()))
	})
}
