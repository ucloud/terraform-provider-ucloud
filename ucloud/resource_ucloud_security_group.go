package ucloud

import (
	"bytes"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecurityGroupCreate,
		Read:   resourceUCloudSecurityGroupRead,
		Update: resourceUCloudSecurityGroupUpdate,
		Delete: resourceUCloudSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "SecurityGroup",
				ValidateFunc: validateSecurityGroupName,
			},

			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port_range": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateSecurityGroupPort,
						},

						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "TCP",
							ValidateFunc: validateStringInChoices([]string{"TCP", "UDP", "GRE", "ICMP"}),
						},

						"cidr_block": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0.0.0.0/0",
							ValidateFunc: validateCidrBlock,
						},

						"policy": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "ACCEPT",
							ValidateFunc: validateStringInChoices([]string{"ACCEPT", "DROP"}),
						},

						"priority": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "HIGH",
							ValidateFunc: validateStringInChoices([]string{"HIGH", "MEDIUM", "LOW"}),
						},
					},
				},
				Set: resourceucloudSecurityGroupRuleHash,
			},

			"tag": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"remark": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).unetconn

	req := conn.NewCreateFirewallRequest()
	req.Name = ucloud.String(d.Get("name").(string))

	if val, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(val.(string))
	}

	req.Rule = buildRuleParameter(d.Get("rules"))

	resp, err := conn.CreateFirewall(req)
	if err != nil {
		return fmt.Errorf("error in create security group, %s", err)
	}

	d.SetId(resp.FWId)

	time.Sleep(3 * time.Second)

	return resourceUCloudSecurityGroupUpdate(d, meta)
}

func resourceUCloudSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).unetconn

	d.Partial(true)

	if d.HasChange("rules") && !d.IsNewResource() {
		d.SetPartial("rules")
		req := conn.NewUpdateFirewallRequest()
		req.FWId = ucloud.String(d.Id())
		req.Rule = buildRuleParameter(d.Get("rules"))
		_, err := conn.UpdateFirewall(req)

		if err != nil {
			return fmt.Errorf("do %s failed in update security group %s, %s", "UpdateFirewall", d.Id(), err)
		}

		time.Sleep(3 * time.Second)
	}

	isChanged := false
	req := conn.NewUpdateFirewallAttributeRequest()
	req.FWId = ucloud.String(d.Id())

	if d.HasChange("name") && !d.IsNewResource() {
		isChanged = true
		req.Name = ucloud.String(d.Get("name").(string))
		d.SetPartial("name")
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("tag").(string))
		d.SetPartial("tag")
	}

	if d.HasChange("remark") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("remark").(string))
		d.SetPartial("remark")
	}

	if isChanged {
		_, err := conn.UpdateFirewallAttribute(req)

		if err != nil {
			return fmt.Errorf("do %s failed in update security group %s, %s", "UpdateFirewallAttribute", d.Id(), err)
		}

		time.Sleep(3 * time.Second)
	}

	d.Partial(false)

	return resourceUCloudSecurityGroupRead(d, meta)
}

func resourceUCloudSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	sgSet, err := client.describeFirewallById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read security group %s, %s", "DescribeFirewall", d.Id(), err)
	}

	d.Set("name", sgSet.Name)
	d.Set("tag", sgSet.Tag)
	d.Set("remark", sgSet.Remark)
	d.Set("create_time", timestampToString(sgSet.CreateTime))

	rules := []map[string]interface{}{}
	for _, item := range sgSet.Rule {
		rules = append(rules, map[string]interface{}{
			"port_range": item.DstPort,
			"protocol":   item.ProtocolType,
			"cidr_block": item.SrcIP,
			"policy":     item.RuleAction,
			"priority":   item.Priority,
		})
	}
	d.Set("rules", rules)

	return nil
}

func resourceUCloudSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	req := conn.NewDeleteFirewallRequest()
	req.FWId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteFirewall(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete security group %s, %s", d.Id(), err))
		}

		_, err := client.describeFirewallById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete security group %s, %s", "DescribeFirewall", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete security group but it still exists"))
	})
}

func resourceucloudSecurityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["port_range"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["protocol"].(string)))

	if m["cidr_block"].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["cidr_block"].(string)))
	}

	if m["policy"].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["policy"].(string)))
	}

	if m["priority"].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["priority"].(string)))
	}

	return hashcode.String(buf.String())
}

func buildRuleParameter(iface interface{}) []string {
	rules := []string{}
	for _, item := range iface.(*schema.Set).List() {
		rule := item.(map[string]interface{})
		s := fmt.Sprintf("%s|%s|%s|%s|%s", rule["protocol"], rule["port_range"], rule["cidr_block"], rule["policy"], rule["priority"])
		rules = append(rules, s)
	}
	return rules
}
