package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/customdiff"
	"github.com/hashicorp/terraform/helper/validation"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudNatGatewayRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudNatGatewayRuleCreate,
		Update: resourceUCloudNatGatewayRuleUpdate,
		Read:   resourceUCloudNatGatewayRuleRead,
		Delete: resourceUCloudNatGatewayRuleDelete,

		CustomizeDiff: customdiff.All(
			diffValidateSrcPortRangeWithDstPortRange,
		),

		Schema: map[string]*schema.Schema{
			"nat_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"tcp",
					"udp",
				}, false),
			},

			//TODO:src_elastic_ip?
			"src_eip_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"src_port_range": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePortRange,
			},
			//TODO:dst_private_ip?
			"dst_ip": {
				Type:     schema.TypeString,
				Required: true,
			},

			"dst_port_range": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePortRange,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateNatGatewayName,
			},
		},
	}
}

func resourceUCloudNatGatewayRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	pro := d.Get("protocol").(string)
	dstIp := d.Get("dst_ip").(string)
	natGwId := d.Get("nat_gateway_id").(string)
	srcEIPId := d.Get("src_eip_id").(string)

	// PreCheck dst_ip
	reqCheckDstIP := conn.NewGetAvailableResourceForPolicyRequest()
	reqCheckDstIP.NATGWId = ucloud.String(natGwId)
	respCheckDstIP, err := conn.GetAvailableResourceForPolicy(reqCheckDstIP)
	if err != nil {
		return fmt.Errorf("error on getting available resource before creating the rule of nat gateway %q, %s", natGwId, err)
	}
	var getAvailableResource bool
	for _, v := range respCheckDstIP.DataSet {
		if v.PrivateIP == dstIp {
			getAvailableResource = true
			break
		}
	}
	if !getAvailableResource {
		return fmt.Errorf("%q is invalid, please get available destination ip for this nat gateway %q", "dst_ip", natGwId)
	}

	// PreCheck scr_eip_id
	respCheckSrcEIPId, err := client.describeNatGatewayById(natGwId)
	if err != nil {
		return fmt.Errorf("error on reading nat gateway %q before creating nat gateway rule, %s", natGwId, err)
	}
	var getAvailableSrcEIPId bool
	for _, v := range respCheckSrcEIPId.IPSet {
		if v.EIPId == srcEIPId {
			getAvailableSrcEIPId = true
			break
		}
	}
	if !getAvailableSrcEIPId {
		return fmt.Errorf("%q is invalid, please get available source eip id for this nat gateway %q", "dst_ip", natGwId)
	}

	//Create nat_gateway_rule
	reqCreate := conn.NewCreateNATGWPolicyRequest()
	reqCreate.NATGWId = ucloud.String(natGwId)
	reqCreate.Protocol = ucloud.String(upperCvt.unconvert(pro))
	reqCreate.SrcEIPId = ucloud.String(srcEIPId)
	reqCreate.SrcPort = ucloud.String(d.Get("src_port_range").(string))
	reqCreate.DstIP = ucloud.String(dstIp)
	reqCreate.DstPort = ucloud.String(d.Get("dst_port_range").(string))

	if v, ok := d.GetOk("name"); ok {
		reqCreate.PolicyName = ucloud.String(v.(string))
	} else {
		reqCreate.PolicyName = ucloud.String(resource.PrefixedUniqueId("tf-nat-gateway-rule-"))
	}

	resp, err := conn.CreateNATGWPolicy(reqCreate)

	if err != nil {
		return fmt.Errorf("error on creating nat gateway rule, %s", err)
	}

	d.SetId(resp.PolicyId)

	return resourceUCloudNatGatewayRuleRead(d, meta)
}

func resourceUCloudNatGatewayRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewUpdateNATGWPolicyRequest()
	req.NATGWId = ucloud.String(d.Get("nat_gateway_id").(string))
	req.PolicyId = ucloud.String(d.Id())
	d.Partial(true)
	if !d.IsNewResource() && (d.HasChange("protocol") || d.HasChange("src_eip_id") || d.HasChange("src_port_range") || d.HasChange("dst_ip") || d.HasChange("dst_port_range") || d.HasChange("name")) {
		req.Protocol = ucloud.String(d.Get("protocol").(string))
		req.SrcEIPId = ucloud.String(d.Get("src_eip_id").(string))
		req.SrcPort = ucloud.String(d.Get("src_port_range").(string))
		req.DstIP = ucloud.String(d.Get("dst_ip").(string))
		req.DstPort = ucloud.String(d.Get("dst_port_range").(string))
		req.PolicyName = ucloud.String(d.Get("name").(string))
		if _, err := conn.UpdateNATGWPolicy(req); err != nil {
			return fmt.Errorf("error on %s to nat_gateway rule %q, %s", "UpdateNATGWPolicy", d.Id(), err)
		}
		d.SetPartial("protocol")
		d.SetPartial("src_eip_id")
		d.SetPartial("src_port_range")
		d.SetPartial("dst_ip")
		d.SetPartial("dst_port_range")
		d.SetPartial("name")
	}

	d.Partial(false)
	return resourceUCloudNatGatewayRuleRead(d, meta)
}

func resourceUCloudNatGatewayRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	policySet, err := client.describeNatGatewayRuleById(d.Id(), d.Get("nat_gateway_id").(string))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading nat gateway rule %q, %s", d.Id(), err)
	}

	d.Set("nat_gateway_id", policySet.NATGWId)
	d.Set("protocol", upperCvt.convert(policySet.Protocol))
	d.Set("src_eip_id", policySet.SrcEIPId)
	d.Set("src_port_range", policySet.SrcPort)
	d.Set("dst_ip", policySet.DstIP)
	d.Set("dst_port_range", policySet.DstPort)
	d.Set("name", policySet.PolicyName)

	return nil
}

func resourceUCloudNatGatewayRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	natGwId := d.Get("nat_gateway_id").(string)
	req := conn.NewDeleteNATGWPolicyRequest()
	req.NATGWId = ucloud.String(natGwId)
	req.PolicyId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNATGWPolicy(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting nat gateway rule %q, %s", d.Id(), err))
		}

		_, err := client.describeNatGatewayRuleById(d.Id(), natGwId)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading nat_gateway rule when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified nat gateway rule %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateSrcPortRangeWithDstPortRange(diff *schema.ResourceDiff, meta interface{}) error {
	srcPortRange := diff.Get("src_port_range").(string)
	dstPortRange := diff.Get("dst_port_range").(string)
	splitedSrc := strings.Split(srcPortRange, "-")
	splitedDrc := strings.Split(dstPortRange, "-")

	if len(splitedSrc) == 2 || len(splitedDrc) == 2 {
		if srcPortRange != dstPortRange {
			return fmt.Errorf("the src_port_range %q must be same as dst_port_range %q when the port mapping use port range not single port", srcPortRange, dstPortRange)
		}
	}

	return nil
}
