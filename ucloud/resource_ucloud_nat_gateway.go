package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"time"
)

func resourceUCloudNatGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudNatGatewayCreate,
		Update: resourceUCloudNatGatewayUpdate,
		Read:   resourceUCloudNatGatewayRead,
		Delete: resourceUCloudNatGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"eip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"security_group": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enable_white_list": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"white_list": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Set:      schema.HashString,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateNatGatewayName,
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
		},
	}
}

func resourceUCloudNatGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewCreateNATGWRequest()
	req.EIPIds = []string{d.Get("eip_id").(string)}
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetworkIds = schemaSetToStringSlice(d.Get("subnet_ids"))
	req.FirewallId = ucloud.String(d.Get("security_group").(string))

	if v, ok := d.GetOk("name"); ok {
		req.NATGWName = ucloud.String(v.(string))
	} else {
		req.NATGWName = ucloud.String(resource.PrefixedUniqueId("tf-nat-gateway-"))
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

	if d.Get("enable_white_list").(bool) {
		req.IfOpen = ucloud.Int(1)
	} else {
		req.IfOpen = ucloud.Int(0)
	}

	resp, err := conn.CreateNATGW(req)
	if err != nil {
		return fmt.Errorf("error on creating nat gateway, %s", err)
	}

	d.SetId(resp.NATGWId)
	return resourceUCloudNatGatewayUpdate(d, meta)
}

func resourceUCloudNatGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	d.Partial(true)

	if d.HasChange("white_list") {
		reqWhite := conn.NewDescribeWhiteListResourceRequest()
		reqWhite.NATGWIds = []string{d.Id()}
		_, err := conn.DescribeWhiteListResource(reqWhite)
		if err != nil {
			return fmt.Errorf("error on reading white list when updating %q, %s", d.Id(), err)
		}

		o, n := d.GetChange("white_list")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		if len(add) > 0 {
			req := conn.NewAddWhiteListResourceRequest()
			req.ResourceIds = interfaceSliceToStringSlice(add)
			req.NATGWId = ucloud.String(d.Id())
			if _, err := conn.AddWhiteListResource(req); err != nil {
				return fmt.Errorf("error on %s to nat gateway %q, %s", "AddWhiteListResource", d.Id(), err)
			}
		}
		if len(remove) > 0 {
			req := conn.NewDeleteWhiteListResourceRequest()
			req.ResourceIds = interfaceSliceToStringSlice(remove)
			req.NATGWId = ucloud.String(d.Id())

			if _, err := conn.DeleteWhiteListResource(req); err != nil {
				if uErr, ok := err.(uerr.Error); !(ok && uErr.Code() == 54002) {
					return fmt.Errorf("error on %s to nat gateway %q, %s", "DeleteWhiteListResource", d.Id(), err)
				}
			}
		}

		d.SetPartial("white_list")
	}

	// update the `enable_white_list` must be after update the `white_list` to insure the service available.
	if d.HasChange("enable_white_list") && !d.IsNewResource() {
		if d.Get("enable_white_list").(bool) {
			req := conn.NewEnableWhiteListRequest()
			req.NATGWId = ucloud.String(d.Id())
			req.IfOpen = ucloud.Int(1)
			if _, err := conn.EnableWhiteList(req); err != nil {
				return fmt.Errorf("error on %s to nat gateway %q, %s", "EnableWhiteList", d.Id(), err)
			}
		} else {
			req := conn.NewEnableWhiteListRequest()
			req.NATGWId = ucloud.String(d.Id())
			req.IfOpen = ucloud.Int(0)
			if _, err := conn.EnableWhiteList(req); err != nil {
				return fmt.Errorf("error on %s to nat gateway %q, %s", "EnableWhiteList", d.Id(), err)
			}
		}
		d.SetPartial("enable_white_list")
	}

	if d.HasChange("subnet_ids") && !d.IsNewResource() {
		req := conn.NewUpdateNATGWSubnetRequest()
		req.NATGWId = ucloud.String(d.Id())
		req.SubnetworkIds = schemaSetToStringSlice(d.Get("subnet_ids"))

		if _, err := conn.UpdateNATGWSubnet(req); err != nil {
			return fmt.Errorf("error on %s to nat gateway %q, %s", "UpdateNATGWSubnet", d.Id(), err)
		}
		d.SetPartial("subnet_ids")
	}

	d.Partial(false)
	return resourceUCloudNatGatewayRead(d, meta)
}

func resourceUCloudNatGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	ngSet, err := client.describeNatGatewayById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading nat gateway %q, %s", d.Id(), err)
	}

	d.Set("vpc_id", ngSet.VPCId)
	d.Set("name", ngSet.NATGWName)
	d.Set("remark", ngSet.Remark)
	d.Set("tag", ngSet.Tag)
	d.Set("create_time", timestampToString(ngSet.CreateTime))
	d.Set("security_group", ngSet.FirewallId)

	var subnetIds []string
	for _, item := range ngSet.SubnetSet {
		subnetIds = append(subnetIds, item.SubnetworkId)
	}
	d.Set("subnet_ids", subnetIds)

	if len(ngSet.IPSet) > 1 {
		var eipIds []string
		for _, item := range ngSet.IPSet {
			eipIds = append(eipIds, item.EIPId)
		}
		return fmt.Errorf("expect only one eip binded to the nat gateway %q, got %v. If you want to bind more than one eip, please manage it through the console or API", d.Id(), eipIds)
	}
	d.Set("eip_id", ngSet.IPSet[0].EIPId)

	req := conn.NewDescribeWhiteListResourceRequest()
	req.NATGWIds = []string{d.Id()}
	whiteSet, err := conn.DescribeWhiteListResource(req)
	if err != nil {
		return fmt.Errorf("error on reading white list when reading nat gateway %q, %s", d.Id(), err)
	}
	var whiteList []string
	for _, v := range whiteSet.DataSet[0].ObjectIPInfo {
		whiteList = append(whiteList, v.ResourceId)
	}
	d.Set("white_list", whiteList)
	if whiteSet.DataSet[0].IfOpen == 1 {
		d.Set("enable_white_list", true)
	} else {
		d.Set("enable_white_list", false)
	}

	return nil
}

func resourceUCloudNatGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteNATGWRequest()
	req.NATGWId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNATGW(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting nat gateway %q, %s", d.Id(), err))
		}

		_, err := client.describeNatGatewayById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading nat gateway when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified nat gateway %q has not been deleted due to unknown error", d.Id()))
	})
}
