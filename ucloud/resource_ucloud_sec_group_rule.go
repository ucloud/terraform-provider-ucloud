package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudSecGroupRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecGroupRuleCreate,
		Read:   resourceUCloudSecGroupRuleRead,
		Update: resourceUCloudSecGroupRuleUpdate,
		Delete: resourceUCloudSecGroupRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"sec_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Ingress",
					"Egress",
				}, false),
			},

			"protocol_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"TCP",
					"UDP",
					"ICMP",
					"ICMPv6",
					"ALL",
				}, false),
			},

			"dst_port": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ip_range": {
				Type:     schema.TypeString,
				Required: true,
			},

			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Accept",
					"Drop",
				}, false),
			},

			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 200),
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceUCloudSecGroupRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	secGroupId := d.Get("sec_group_id").(string)

	// the rules of one sec group are written one by one, serialize them so
	// that parallel applies do not race on the same sec group
	ucloudMutexKV.Lock(secGroupId)
	defer ucloudMutexKV.Unlock(secGroupId)

	req := conn.NewCreateSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupId)
	req.Rule = []vpc.CreateSecGroupRuleParamRule{
		{
			Direction:    ucloud.String(d.Get("direction").(string)),
			ProtocolType: ucloud.String(d.Get("protocol_type").(string)),
			DstPort:      ucloud.String(d.Get("dst_port").(string)),
			IPRange:      ucloud.String(d.Get("ip_range").(string)),
			RuleAction:   ucloud.String(d.Get("rule_action").(string)),
			Priority:     ucloud.Int(d.Get("priority").(int)),
			Remark:       ucloud.String(d.Get("remark").(string)),
		},
	}

	resp, err := conn.CreateSecGroupRule(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group rule for %q, %s", secGroupId, err)
	}

	if len(resp.RuleId) < 1 {
		return fmt.Errorf("error on creating sec group rule for %q, no rule id returned", secGroupId)
	}

	d.SetId(buildUCloudSecGroupRuleID(secGroupId, resp.RuleId[0]))

	return resourceUCloudSecGroupRuleRead(d, meta)
}

func resourceUCloudSecGroupRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	secGroupId, ruleId, err := parseUCloudSecGroupRuleID(d.Id())
	if err != nil {
		return err
	}

	ucloudMutexKV.Lock(secGroupId)
	defer ucloudMutexKV.Unlock(secGroupId)

	// UpdateSecGroupRule requires every field of the rule, so the whole rule is
	// sent on any change instead of a per field diff
	req := conn.NewUpdateSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupId)
	req.Rule = []vpc.UpdateSecGroupRuleParamRule{
		{
			RuleId:       ucloud.String(ruleId),
			Direction:    ucloud.String(d.Get("direction").(string)),
			ProtocolType: ucloud.String(d.Get("protocol_type").(string)),
			DstPort:      ucloud.String(d.Get("dst_port").(string)),
			IPRange:      ucloud.String(d.Get("ip_range").(string)),
			RuleAction:   ucloud.String(d.Get("rule_action").(string)),
			Priority:     ucloud.Int(d.Get("priority").(int)),
			Remark:       ucloud.String(d.Get("remark").(string)),
		},
	}

	if _, err := conn.UpdateSecGroupRule(req); err != nil {
		return fmt.Errorf("error on %s to sec group rule %q, %s", "UpdateSecGroupRule", d.Id(), err)
	}

	return resourceUCloudSecGroupRuleRead(d, meta)
}

func resourceUCloudSecGroupRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	secGroupId, ruleId, err := parseUCloudSecGroupRuleID(d.Id())
	if err != nil {
		return err
	}

	ruleSet, err := client.describeSecGroupRuleById(secGroupId, ruleId)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sec group rule %q, %s", d.Id(), err)
	}

	d.Set("sec_group_id", secGroupId)
	d.Set("direction", ruleSet.Direction)
	d.Set("protocol_type", ruleSet.ProtocolType)
	d.Set("dst_port", ruleSet.DstPort)
	d.Set("ip_range", ruleSet.IPRange)
	d.Set("rule_action", ruleSet.RuleAction)
	d.Set("priority", ruleSet.Priority)
	d.Set("remark", ruleSet.Remark)

	return nil
}

func resourceUCloudSecGroupRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	secGroupId, ruleId, err := parseUCloudSecGroupRuleID(d.Id())
	if err != nil {
		return err
	}

	ucloudMutexKV.Lock(secGroupId)
	defer ucloudMutexKV.Unlock(secGroupId)

	req := conn.NewDeleteSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupId)
	req.RuleId = []string{ruleId}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteSecGroupRule(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting sec group rule %q, %s", d.Id(), err))
		}

		_, err := client.describeSecGroupRuleById(secGroupId, ruleId)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sec group rule when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sec group rule %q has not been deleted due to unknown error", d.Id()))
	})
}

const ucloudSecGroupRuleIDSeparator = ":"

func buildUCloudSecGroupRuleID(secGroupId, ruleId string) string {
	return strings.Join([]string{secGroupId, ruleId}, ucloudSecGroupRuleIDSeparator)
}

func parseUCloudSecGroupRuleID(id string) (secGroupId string, ruleId string, err error) {
	items := strings.Split(id, ucloudSecGroupRuleIDSeparator)
	if len(items) != 2 || items[0] == "" || items[1] == "" {
		return "", "", fmt.Errorf(
			"invalid sec group rule id %q, expected %q",
			id, "<sec_group_id>"+ucloudSecGroupRuleIDSeparator+"<rule_id>",
		)
	}

	return items[0], items[1], nil
}
