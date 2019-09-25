package ucloud

import (
	"bytes"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/customdiff"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

// security policy use ICMP, GRE packet with port is not supported
var portIndependentProtocols = []string{"icmp", "gre"}

func resourceUCloudSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecurityGroupCreate,
		Read:   resourceUCloudSecurityGroupRead,
		Update: resourceUCloudSecurityGroupUpdate,
		Delete: resourceUCloudSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(
			diffValidatePortRangeWithProtocol,
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validatePortRange,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v, ok := d.GetOk("protocol"); ok && shouldIgnorePort(v.(string)) {
									return true
								}
								return false
							},
						},

						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "tcp",
							ValidateFunc: validation.StringInSlice([]string{
								"tcp",
								"udp",
								"gre",
								"icmp",
							}, false),
						},

						"cidr_block": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0.0.0.0/0",
							ValidateFunc: validation.CIDRNetwork(0, 32),
						},

						"policy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "accept",
							ValidateFunc: validation.StringInSlice([]string{
								"accept",
								"drop",
							}, false),
						},

						"priority": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "high",
							ValidateFunc: validation.StringInSlice([]string{
								"high",
								"medium",
								"low",
							}, false),
						},
					},
				},
				Set: resourceucloudSecurityGroupRuleHash,
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

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	req := conn.NewCreateFirewallRequest()
	req.Rule = buildRuleParameter(d.Get("rules"))

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-security-group-"))
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

	resp, err := conn.CreateFirewall(req)
	if err != nil {
		return fmt.Errorf("error on creating security group, %s", err)
	}

	d.SetId(resp.FWId)

	// after create security group, we need to wait it initialized
	stateConf := securityWaitForState(client, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for security group %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudSecurityGroupRead(d, meta)
}

func resourceUCloudSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	d.Partial(true)

	if d.HasChange("rules") && !d.IsNewResource() {
		req := conn.NewUpdateFirewallRequest()
		req.FWId = ucloud.String(d.Id())
		req.Rule = buildRuleParameter(d.Get("rules"))
		_, err := conn.UpdateFirewall(req)

		if err != nil {
			return fmt.Errorf("error on %s to security group %q, %s", "UpdateFirewall", d.Id(), err)
		}

		d.SetPartial("rules")

		// after update security group rule, we need to wait it completed
		stateConf := securityWaitForState(client, d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error on waiting for %s complete to security group %q, %s", "UpdateFirewall", d.Id(), err)
		}
	}

	isChanged := false
	req := conn.NewUpdateFirewallAttributeRequest()
	req.FWId = ucloud.String(d.Id())

	if d.HasChange("name") && !d.IsNewResource() {
		isChanged = true
		req.Name = ucloud.String(d.Get("name").(string))
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		isChanged = true

		// if tag is empty string, use default tag
		if v, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(v.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}
	}

	if d.HasChange("remark") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("remark").(string))
	}

	if isChanged {
		_, err := conn.UpdateFirewallAttribute(req)
		if err != nil {
			return fmt.Errorf("error on %s to security group %q, %s", "UpdateFirewallAttribute", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")

		// after update security group attribute, we need to wait it completed
		stateConf := securityWaitForState(client, d.Id())
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error on waiting for %s complete to security group %q, %s", "UpdateFirewallAttribute", d.Id(), err)
		}
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
		return fmt.Errorf("error on reading security group %q, %s", d.Id(), err)
	}

	d.Set("name", sgSet.Name)
	d.Set("tag", sgSet.Tag)
	d.Set("remark", sgSet.Remark)
	d.Set("create_time", timestampToString(sgSet.CreateTime))

	rules := []map[string]interface{}{}
	for _, item := range sgSet.Rule {
		rules = append(rules, map[string]interface{}{
			"port_range": item.DstPort,
			"protocol":   upperCvt.convert(item.ProtocolType),
			"cidr_block": item.SrcIP,
			"policy":     upperCvt.convert(item.RuleAction),
			"priority":   upperCvt.convert(item.Priority),
		})
	}

	if err := d.Set("rules", rules); err != nil {
		return err
	}

	return nil
}

func resourceUCloudSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	req := conn.NewDeleteFirewallRequest()
	req.FWId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteFirewall(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting security group %q, %s", d.Id(), err))
		}

		_, err := client.describeFirewallById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading security group when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified security group %q has not been deleted due to unknown error", d.Id()))
	})
}

func resourceucloudSecurityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	protocol := m["protocol"].(string)
	if !shouldIgnorePort(protocol) {
		buf.WriteString(fmt.Sprintf("%s-", m["port_range"].(string)))
	}

	buf.WriteString(fmt.Sprintf("%s-", protocol))

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
		port := rule["port_range"]
		if v := rule["protocol"].(string); shouldIgnorePort(v) {
			port = ""
		}
		s := fmt.Sprintf(
			"%s|%s|%s|%s|%s",
			upperCvt.unconvert(rule["protocol"].(string)),
			port,
			rule["cidr_block"],
			upperCvt.unconvert(rule["policy"].(string)),
			upperCvt.unconvert(rule["priority"].(string)),
		)
		rules = append(rules, s)
	}
	return rules
}

func securityWaitForState(client *UCloudClient, sgId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			sgSet, err := client.describeFirewallById(sgId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			return sgSet, statusInitialized, nil
		},
	}
}

func shouldIgnorePort(protocol string) bool {
	return checkStringIn(protocol, portIndependentProtocols) == nil
}

func diffValidatePortRangeWithProtocol(diff *schema.ResourceDiff, v interface{}) error {
	for _, item := range diff.Get("rules").(*schema.Set).List() {
		rule := item.(map[string]interface{})

		if v := rule["protocol"].(string); !shouldIgnorePort(v) && rule["port_range"].(string) == "" {
			return fmt.Errorf("%q must be set when %q is %q or %q", "port_range", "protocol", "tcp", "udp")
		}
	}
	return nil
}
