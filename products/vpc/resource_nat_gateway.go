package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudNatGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating nat gateway, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewCreateNATGWRequest()
	req.EIPIds = []string{d.Get("eip_id").(string)}
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetworkIds = schemaSetToStringSlice(d.Get("subnet_ids"))
	req.FirewallId = ucloud.String(d.Get("security_group").(string))
	if value, ok := d.GetOk("name"); ok {
		req.NATGWName = ucloud.String(value.(string))
	} else {
		req.NATGWName = ucloud.String(resource.PrefixedUniqueId("tf-nat-gateway-"))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}
	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating nat gateway, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	if d.HasChange("security_group") && !d.IsNewResource() {
		request := client.unetconn.NewGrantFirewallRequest()
		request.FWId = ucloud.String(d.Get("security_group").(string))
		request.ResourceType = ucloud.String(eipResourceTypeNatGateway)
		request.ResourceId = ucloud.String(d.Id())
		if _, err := client.unetconn.GrantFirewall(request); err != nil {
			return fmt.Errorf("error on %s to nat gateway %q, %s", "GrantFirewall", d.Id(), err)
		}
		d.SetPartial("security_group")
	}

	if d.HasChange("white_list") {
		request := conn.NewDescribeWhiteListResourceRequest()
		request.NATGWIds = []string{d.Id()}
		if _, err := conn.DescribeWhiteListResource(request); err != nil {
			return fmt.Errorf("error on reading white list when updating %q, %s", d.Id(), err)
		}
		oldValue, newValue := d.GetChange("white_list")
		oldSet, newSet := oldValue.(*schema.Set), newValue.(*schema.Set)
		remove := oldSet.Difference(newSet).List()
		add := newSet.Difference(oldSet).List()
		if len(add) > 0 {
			request := conn.NewAddWhiteListResourceRequest()
			request.ResourceIds = interfaceSliceToStringSlice(add)
			request.NATGWId = ucloud.String(d.Id())
			if _, err := conn.AddWhiteListResource(request); err != nil {
				return fmt.Errorf("error on %s to nat gateway %q, %s", "AddWhiteListResource", d.Id(), err)
			}
		}
		if len(remove) > 0 {
			request := conn.NewDeleteWhiteListResourceRequest()
			request.ResourceIds = interfaceSliceToStringSlice(remove)
			request.NATGWId = ucloud.String(d.Id())
			if _, err := conn.DeleteWhiteListResource(request); err != nil {
				if cloudErr, ok := err.(uerr.Error); !(ok && cloudErr.Code() == 54002) {
					return fmt.Errorf("error on %s to nat gateway %q, %s", "DeleteWhiteListResource", d.Id(), err)
				}
			}
		}
		d.SetPartial("white_list")
	}

	// Keep whitelist membership changes ahead of toggling the service state.
	if d.HasChange("enable_white_list") && !d.IsNewResource() {
		request := conn.NewEnableWhiteListRequest()
		request.NATGWId = ucloud.String(d.Id())
		if d.Get("enable_white_list").(bool) {
			request.IfOpen = ucloud.Int(1)
		} else {
			request.IfOpen = ucloud.Int(0)
		}
		if _, err := conn.EnableWhiteList(request); err != nil {
			return fmt.Errorf("error on %s to nat gateway %q, %s", "EnableWhiteList", d.Id(), err)
		}
		d.SetPartial("enable_white_list")
	}

	if d.HasChange("subnet_ids") && !d.IsNewResource() {
		request := conn.NewUpdateNATGWSubnetRequest()
		request.NATGWId = ucloud.String(d.Id())
		request.SubnetworkIds = schemaSetToStringSlice(d.Get("subnet_ids"))
		if _, err := conn.UpdateNATGWSubnet(request); err != nil {
			return fmt.Errorf("error on %s to nat gateway %q, %s", "UpdateNATGWSubnet", d.Id(), err)
		}
		d.SetPartial("subnet_ids")
	}
	d.Partial(false)
	return resourceUCloudNatGatewayRead(d, meta)
}

func resourceUCloudNatGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading nat gateway, %s", err)
	}
	conn := client.vpcconn
	natGateway, err := client.describeNatGatewayById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading nat gateway %q, %s", d.Id(), err)
	}
	_ = d.Set("vpc_id", natGateway.VPCId)
	_ = d.Set("name", natGateway.NATGWName)
	_ = d.Set("remark", natGateway.Remark)
	_ = d.Set("tag", natGateway.Tag)
	_ = d.Set("create_time", timestampToString(natGateway.CreateTime))
	_ = d.Set("security_group", natGateway.FirewallId)
	subnetIDs := []string{}
	for _, item := range natGateway.SubnetSet {
		subnetIDs = append(subnetIDs, item.SubnetworkId)
	}
	_ = d.Set("subnet_ids", subnetIDs)
	if len(natGateway.IPSet) > 1 {
		eipIDs := []string{}
		for _, item := range natGateway.IPSet {
			eipIDs = append(eipIDs, item.EIPId)
		}
		return fmt.Errorf("expect only one eip binded to the nat gateway %q, got %v. If you want to bind more than one eip, please manage it through the console or API", d.Id(), eipIDs)
	}
	_ = d.Set("eip_id", natGateway.IPSet[0].EIPId)
	request := conn.NewDescribeWhiteListResourceRequest()
	request.NATGWIds = []string{d.Id()}
	whiteSet, err := conn.DescribeWhiteListResource(request)
	if err != nil {
		return fmt.Errorf("error on reading white list when reading nat gateway %q, %s", d.Id(), err)
	}
	whiteList := []string{}
	for _, item := range whiteSet.DataSet[0].ObjectIPInfo {
		whiteList = append(whiteList, item.ResourceId)
	}
	_ = d.Set("white_list", whiteList)
	_ = d.Set("enable_white_list", whiteSet.DataSet[0].IfOpen == 1)
	return nil
}

func resourceUCloudNatGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting nat gateway, %s", err)
	}
	conn := client.vpcconn
	request := conn.NewDeleteNATGWRequest()
	request.NATGWId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNATGW(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting nat gateway %q, %s", d.Id(), err))
		}
		if _, err := client.describeNatGatewayById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading nat gateway when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified nat gateway %q has not been deleted due to unknown error", d.Id()))
	})
}
