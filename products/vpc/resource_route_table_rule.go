package vpc

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudRouteTableRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudRouteTableRuleCreate,
		Update: resourceUCloudRouteTableRuleUpdate,
		Read:   resourceUCloudRouteTableRuleRead,
		Delete: resourceUCloudRouteTableRuleDelete,
		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dst_addr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRouteRuleDstAddr,
			},
			"nexthop_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"instance",
					"vip",
				}, false),
			},
			"nexthop_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUCloudRouteTableRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating route table rule, %s", err)
	}
	conn := client.vpcconn

	routeTableID := d.Get("route_table_id").(string)
	dstAddr := d.Get("dst_addr").(string)
	nexthopType := d.Get("nexthop_type").(string)
	nexthopID := d.Get("nexthop_id").(string)
	remark := d.Get("remark").(string)

	req := conn.NewModifyRouteRuleRequest()
	req.RouteTableId = ucloud.String(routeTableID)
	// For "add" the route rule id is a placeholder; the real id is assigned by
	// the API and discovered afterwards through DescribeRouteTable.
	placeholderID := resource.PrefixedUniqueId("tf-route-rule-")
	req.RouteRule = []string{buildRouteRule(placeholderID, dstAddr, nexthopType, nexthopID, remark, "add")}

	if _, err := conn.ModifyRouteRule(req); err != nil {
		return fmt.Errorf("error on creating route table rule, %s", err)
	}

	rule, err := client.describeRouteRuleByAttrs(routeTableID, dstAddr, nexthopID)
	if err != nil {
		return fmt.Errorf("error on reading route table rule after creating, %s", err)
	}
	d.SetId(rule.RouteRuleId)
	return resourceUCloudRouteTableRuleRead(d, meta)
}

func resourceUCloudRouteTableRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating route table rule, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	if !d.IsNewResource() && (d.HasChange("nexthop_id") || d.HasChange("remark")) {
		req := conn.NewModifyRouteRuleRequest()
		req.RouteTableId = ucloud.String(d.Get("route_table_id").(string))
		req.RouteRule = []string{buildRouteRule(
			d.Id(),
			d.Get("dst_addr").(string),
			d.Get("nexthop_type").(string),
			d.Get("nexthop_id").(string),
			d.Get("remark").(string),
			"update",
		)}
		if _, err := conn.ModifyRouteRule(req); err != nil {
			return fmt.Errorf("error on %s to route table rule %q, %s", "ModifyRouteRule", d.Id(), err)
		}
		d.SetPartial("nexthop_id")
		d.SetPartial("remark")
	}
	d.Partial(false)
	return resourceUCloudRouteTableRuleRead(d, meta)
}

func resourceUCloudRouteTableRuleRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading route table rule, %s", err)
	}
	rule, err := client.describeRouteRuleById(d.Get("route_table_id").(string), d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading route table rule %q, %s", d.Id(), err)
	}
	_ = d.Set("route_table_id", rule.RouteTableId)
	_ = d.Set("dst_addr", rule.DstAddr)
	_ = d.Set("nexthop_type", strings.ToLower(rule.NexthopType))
	_ = d.Set("nexthop_id", rule.NexthopId)
	_ = d.Set("remark", rule.Remark)
	return nil
}

func resourceUCloudRouteTableRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting route table rule, %s", err)
	}
	conn := client.vpcconn
	routeTableID := d.Get("route_table_id").(string)
	req := conn.NewModifyRouteRuleRequest()
	req.RouteTableId = ucloud.String(routeTableID)
	req.RouteRule = []string{buildRouteRule(
		d.Id(),
		d.Get("dst_addr").(string),
		d.Get("nexthop_type").(string),
		d.Get("nexthop_id").(string),
		d.Get("remark").(string),
		"delete",
	)}
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.ModifyRouteRule(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting route table rule %q, %s", d.Id(), err))
		}
		if _, err := client.describeRouteRuleById(routeTableID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading route table rule when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified route table rule %q has not been deleted due to unknown error", d.Id()))
	})
}

// buildRouteRule renders a single route rule in the pipe-separated format the
// ModifyRouteRule API expects:
//
//	RouteRuleId | DstAddr | NexthopType | NexthopId | Priority | Remark | Action
func buildRouteRule(routeRuleID, dstAddr, nexthopType, nexthopID, remark, action string) string {
	return strings.Join([]string{routeRuleID, dstAddr, nexthopType, nexthopID, "0", remark, action}, "|")
}
