package vpc

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"src_eip_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"src_port_range": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePortRange,
			},
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating nat gateway rule, %s", err)
	}
	conn := client.vpcconn
	protocol := d.Get("protocol").(string)
	dstIP := d.Get("dst_ip").(string)
	natGatewayID := d.Get("nat_gateway_id").(string)
	srcEIPID := d.Get("src_eip_id").(string)

	checkDstIP := conn.NewGetAvailableResourceForPolicyRequest()
	checkDstIP.NATGWId = ucloud.String(natGatewayID)
	available, err := conn.GetAvailableResourceForPolicy(checkDstIP)
	if err != nil {
		return fmt.Errorf("error on getting available resource before creating the rule of nat gateway %q, %s", natGatewayID, err)
	}
	availableDstIP := false
	for _, item := range available.DataSet {
		if item.PrivateIP == dstIP {
			availableDstIP = true
			break
		}
	}
	if !availableDstIP {
		return fmt.Errorf("%q is invalid, please get available destination ip for this nat gateway %q", "dst_ip", natGatewayID)
	}

	natGateway, err := client.describeNatGatewayById(natGatewayID)
	if err != nil {
		return fmt.Errorf("error on reading nat gateway %q before creating nat gateway rule, %s", natGatewayID, err)
	}
	availableSrcEIPID := false
	for _, item := range natGateway.IPSet {
		if item.EIPId == srcEIPID {
			availableSrcEIPID = true
			break
		}
	}
	if !availableSrcEIPID {
		return fmt.Errorf("%q is invalid, please get available source eip id for this nat gateway %q", "dst_ip", natGatewayID)
	}

	request := conn.NewCreateNATGWPolicyRequest()
	request.NATGWId = ucloud.String(natGatewayID)
	request.Protocol = ucloud.String(upperCvt.unconvert(protocol))
	request.SrcEIPId = ucloud.String(srcEIPID)
	request.SrcPort = ucloud.String(d.Get("src_port_range").(string))
	request.DstIP = ucloud.String(dstIP)
	request.DstPort = ucloud.String(d.Get("dst_port_range").(string))
	if value, ok := d.GetOk("name"); ok {
		request.PolicyName = ucloud.String(value.(string))
	} else {
		request.PolicyName = ucloud.String(resource.PrefixedUniqueId("tf-nat-gateway-rule-"))
	}
	resp, err := conn.CreateNATGWPolicy(request)
	if err != nil {
		return fmt.Errorf("error on creating nat gateway rule, %s", err)
	}
	d.SetId(resp.PolicyId)
	return resourceUCloudNatGatewayRuleRead(d, meta)
}

func resourceUCloudNatGatewayRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating nat gateway rule, %s", err)
	}
	conn := client.vpcconn
	request := conn.NewUpdateNATGWPolicyRequest()
	request.NATGWId = ucloud.String(d.Get("nat_gateway_id").(string))
	request.PolicyId = ucloud.String(d.Id())
	d.Partial(true)
	if !d.IsNewResource() && (d.HasChange("protocol") || d.HasChange("src_eip_id") || d.HasChange("src_port_range") || d.HasChange("dst_ip") || d.HasChange("dst_port_range") || d.HasChange("name")) {
		request.Protocol = ucloud.String(d.Get("protocol").(string))
		request.SrcEIPId = ucloud.String(d.Get("src_eip_id").(string))
		request.SrcPort = ucloud.String(d.Get("src_port_range").(string))
		request.DstIP = ucloud.String(d.Get("dst_ip").(string))
		request.DstPort = ucloud.String(d.Get("dst_port_range").(string))
		request.PolicyName = ucloud.String(d.Get("name").(string))
		if _, err := conn.UpdateNATGWPolicy(request); err != nil {
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading nat gateway rule, %s", err)
	}
	policy, err := client.describeNatGatewayRuleById(d.Id(), d.Get("nat_gateway_id").(string))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading nat gateway rule %q, %s", d.Id(), err)
	}
	_ = d.Set("nat_gateway_id", policy.NATGWId)
	_ = d.Set("protocol", upperCvt.convert(policy.Protocol))
	_ = d.Set("src_eip_id", policy.SrcEIPId)
	_ = d.Set("src_port_range", policy.SrcPort)
	_ = d.Set("dst_ip", policy.DstIP)
	_ = d.Set("dst_port_range", policy.DstPort)
	_ = d.Set("name", policy.PolicyName)
	return nil
}

func resourceUCloudNatGatewayRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting nat gateway rule, %s", err)
	}
	conn := client.vpcconn
	natGatewayID := d.Get("nat_gateway_id").(string)
	request := conn.NewDeleteNATGWPolicyRequest()
	request.NATGWId = ucloud.String(natGatewayID)
	request.PolicyId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNATGWPolicy(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting nat gateway rule %q, %s", d.Id(), err))
		}
		if _, err := client.describeNatGatewayRuleById(d.Id(), natGatewayID); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading nat_gateway rule when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified nat gateway rule %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateSrcPortRangeWithDstPortRange(diff *schema.ResourceDiff, _ interface{}) error {
	srcPortRange := diff.Get("src_port_range").(string)
	dstPortRange := diff.Get("dst_port_range").(string)
	srcParts := strings.Split(srcPortRange, "-")
	dstParts := strings.Split(dstPortRange, "-")
	if len(srcParts) == 2 || len(dstParts) == 2 {
		if srcPortRange != dstPortRange {
			return fmt.Errorf("the src_port_range %q must be same as dst_port_range %q when the port mapping use port range not single port", srcPortRange, dstPortRange)
		}
	}
	return nil
}
