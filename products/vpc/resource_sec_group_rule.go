package vpc

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

// secGroupPortIndependentProtocols lists the protocols a destination port
// carries no meaning for. dst_port only applies to the other two, which keeps
// one rule exactly one representation: nothing is sent for these protocols,
// so whatever the API echoes back cannot disagree with the configuration and
// produce a diff on every plan.
var secGroupPortIndependentProtocols = []string{"ICMP", "ICMPv6", "ALL"}

// secGroupRulePortAll is the destination port a rule that has no port of its own
// is sent with. It is a destination port here, not a protocol, even though it is
// spelled like one of the entries above. The API stores an empty port but names
// it "ALL" on the wire, and an empty one cannot travel at all: the SDK's form
// encoder drops empty values before the request leaves, and the API reads the
// resulting missing parameter as an invalid rule rather than as "all ports" - it
// rejects the call with 208605.
const secGroupRulePortAll = "ALL"

// resourceUCloudSecGroupRule manages one rule of a security group. The rule is
// its own resource rather than a block inside ucloud_sec_group because the API
// gives every rule a RuleId of its own: addressing a rule by that ID is what
// lets a change stay in place instead of replacing the rule, and it is what
// makes a single rule importable, targetable and movable on its own.
func resourceUCloudSecGroupRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecGroupRuleCreate,
		Read:   resourceUCloudSecGroupRuleRead,
		Update: resourceUCloudSecGroupRuleUpdate,
		Delete: resourceUCloudSecGroupRuleDelete,
		Importer: &schema.ResourceImporter{
			State: importUCloudSecGroupRule,
		},

		// A destination port is required for TCP and UDP and meaningless for the
		// rest. A ValidateFunc sees one field at a time, so the check needs the
		// whole configuration and has to happen here.
		CustomizeDiff: customdiff.All(
			diffValidateSecGroupRulePort,
		),

		Schema: map[string]*schema.Schema{
			// The group is the rule's parent rather than one of its attributes:
			// a rule cannot move between groups, so changing this rebuilds it.
			"sec_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSecGroupResourceID,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSecGroupPortList,
			},

			"ip_range": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateSecGroupIPRange,
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
			},
		},
	}
}

func resourceUCloudSecGroupRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating sec group rule, %s", err)
	}
	conn := client.vpcconn

	secGroupID := d.Get("sec_group_id").(string)

	// CreateSecGroupRule is a batch API, but a rule resource owns exactly one
	// rule, so it is always called with a single element. Sending one also
	// removes any assumption about the order the returned IDs come back in.
	req := conn.NewCreateSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupID)
	req.Rule = []vpcapi.CreateSecGroupRuleParamRule{secGroupRuleParameter(d)}

	resp, err := conn.CreateSecGroupRule(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group rule for sec group %q, %s", secGroupID, err)
	}

	// One rule was sent, so exactly one ID has to come back. Settling for the
	// first of several would leave a rule behind that nothing manages, which is
	// an open door on the network rather than a cosmetic problem. The IDs are
	// deliberately not cleaned up on the way out: a response this far off the
	// contract is not one to delete resources on the strength of, so the error
	// says what to go and look at instead.
	ruleIDCount := 0
	if resp != nil {
		ruleIDCount = len(resp.RuleId)
	}
	if ruleIDCount != 1 {
		return fmt.Errorf("error on creating sec group rule for sec group %q, expected exactly 1 rule id, got %d; check the group for rules this resource does not manage", secGroupID, ruleIDCount)
	}
	if resp.RuleId[0] == "" {
		return fmt.Errorf("error on creating sec group rule for sec group %q, the API returned an empty rule id; check the group for rules this resource does not manage", secGroupID)
	}
	d.SetId(resp.RuleId[0])

	if _, err := secGroupRuleWaitForState(client, secGroupID, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for sec group rule %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudSecGroupRuleRead(d, meta)
}

func resourceUCloudSecGroupRuleRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading sec group rule, %s", err)
	}

	rule, err := client.describeSecGroupRuleById(d.Get("sec_group_id").(string), d.Id())
	if err != nil {
		if isNotFoundError(err) {
			// Either the rule was removed out of band or its group was deleted,
			// which takes the rules with it.
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sec group rule %q, %s", d.Id(), err)
	}

	_ = d.Set("direction", rule.Direction)
	_ = d.Set("protocol_type", rule.ProtocolType)
	_ = d.Set("dst_port", secGroupRuleDstPort(*rule))
	_ = d.Set("ip_range", rule.IPRange)
	_ = d.Set("rule_action", rule.RuleAction)
	_ = d.Set("priority", rule.Priority)
	_ = d.Set("remark", rule.Remark)

	return nil
}

func resourceUCloudSecGroupRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating sec group rule, %s", err)
	}
	conn := client.vpcconn

	secGroupID := d.Get("sec_group_id").(string)

	// UpdateSecGroupRule replaces every attribute of the rule it is given, so
	// the whole rule is sent rather than only the changed fields. The rule is
	// addressed by RuleId alone, which is what keeps an edit in place instead
	// of replacing the rule and handing it a new ID.
	base := secGroupRuleParameter(d)
	req := conn.NewUpdateSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupID)
	req.Rule = []vpcapi.UpdateSecGroupRuleParamRule{{
		Direction:    base.Direction,
		DstPort:      base.DstPort,
		IPRange:      base.IPRange,
		Priority:     base.Priority,
		ProtocolType: base.ProtocolType,
		Remark:       base.Remark,
		RuleAction:   base.RuleAction,
		RuleId:       ucloud.String(d.Id()),
	}}
	if _, err := conn.UpdateSecGroupRule(req); err != nil {
		return fmt.Errorf("error on %s to sec group rule %q, %s", "UpdateSecGroupRule", d.Id(), err)
	}

	// Kept even though it cannot tell "not visible yet" apart from "visible"
	// here: the rule was already there, so it returns as soon as the first read
	// succeeds. What it is for is the replication lag between the primary a
	// write lands on and the replica a read can be served from - the delay
	// gives the write a moment to propagate before the read-back below. It is
	// worth the wait, but it guarantees nothing on its own.
	if _, err := secGroupRuleWaitForState(client, secGroupID, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for sec group rule %q after updating, %s", d.Id(), err)
	}

	return resourceUCloudSecGroupRuleRead(d, meta)
}

func resourceUCloudSecGroupRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting sec group rule, %s", err)
	}
	conn := client.vpcconn

	secGroupID := d.Get("sec_group_id").(string)

	req := conn.NewDeleteSecGroupRuleRequest()
	req.SecGroupId = ucloud.String(secGroupID)
	req.RuleId = []string{d.Id()}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		// A rule that is already gone makes the delete a success rather than an
		// error, so repeated destroys stay idempotent. A missing group counts
		// too, because it took its rules with it.
		if _, err := client.describeSecGroupRuleById(secGroupID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sec group rule when deleting %q, %s", d.Id(), err))
		}

		if _, err := conn.DeleteSecGroupRule(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting sec group rule %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sec group rule %q has not been deleted due to unknown error", d.Id()))
	})
}

// importUCloudSecGroupRule splits the composite import ID. A rule cannot be
// located from its own ID alone, because DescribeSecGroup has no RuleId filter,
// so the parent group comes along: <sec_group_id>/<rule_id>.
func importUCloudSecGroupRule(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	parts := strings.Split(importID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid sec group rule import ID %q: expected <sec_group_id>/<rule_id>", importID)
	}

	secGroupID := strings.TrimSpace(parts[0])
	ruleID := strings.TrimSpace(parts[1])
	if secGroupID == "" || ruleID == "" {
		return nil, fmt.Errorf("invalid sec group rule import ID %q: expected <sec_group_id>/<rule_id>", importID)
	}

	// The pair is checked here rather than left to the read that follows it. A
	// mistyped group ID is well formed, so the import would succeed and the
	// read would then find nothing and clear the ID, leaving the user with an
	// empty state and no idea which half was wrong.
	client, err := clientFromMeta(meta)
	if err != nil {
		return nil, fmt.Errorf("error on getting client when importing sec group rule %q, %s", importID, err)
	}
	if _, err := client.describeSecGroupRuleById(secGroupID, ruleID); err != nil {
		if isNotFoundError(err) {
			return nil, fmt.Errorf("cannot import sec group rule %q: sec group %q has no rule %q", importID, secGroupID, ruleID)
		}
		return nil, fmt.Errorf("error on importing sec group rule %q, %s", importID, err)
	}

	if err := d.Set("sec_group_id", secGroupID); err != nil {
		return nil, fmt.Errorf("import sec group rule %q: %w", importID, err)
	}
	d.SetId(ruleID)

	return []*schema.ResourceData{d}, nil
}

// secGroupRuleParameter reads the rule's attributes out of the configuration.
// Create and update take structurally identical parameters, so this serves both
// and the update path only adds the RuleId it addresses the rule by.
func secGroupRuleParameter(d *schema.ResourceData) vpcapi.CreateSecGroupRuleParamRule {
	// IPVersion is left unset on purpose. The SDK marks it as on its way out
	// once IPv6 is supported, and the API does not ask for it.
	return vpcapi.CreateSecGroupRuleParamRule{
		Direction:    ucloud.String(d.Get("direction").(string)),
		DstPort:      ucloud.String(secGroupRuleDstPortForRequest(d.Get("dst_port").(string))),
		IPRange:      ucloud.String(d.Get("ip_range").(string)),
		Priority:     ucloud.Int(d.Get("priority").(int)),
		ProtocolType: ucloud.String(d.Get("protocol_type").(string)),
		RuleAction:   ucloud.String(d.Get("rule_action").(string)),
		Remark:       ucloud.String(d.Get("remark").(string)),
	}
}

// secGroupRuleDstPortForRequest reports the destination port to send for a rule.
// It is the write side of the pair secGroupRuleDstPort forms the read side of:
// the configuration leaves the port unset on a rule a port carries no meaning
// for, and the API both takes and reports the same rule as having the port that
// stands for the empty one, so the two spellings meet in the middle of the trip.
func secGroupRuleDstPortForRequest(dstPort string) string {
	if dstPort == "" {
		return secGroupRulePortAll
	}
	return dstPort
}

// secGroupRuleDstPort reports the destination port to record for a rule the API
// returned. The protocols a port carries no meaning for are normalized to the
// empty string the configuration uses, so a value the API fills in on its own
// cannot turn into a diff on every plan.
func secGroupRuleDstPort(rule vpcapi.SecGroupRuleInfo) string {
	if isStringIn(rule.ProtocolType, secGroupPortIndependentProtocols) {
		return ""
	}
	return rule.DstPort
}

func diffValidateSecGroupRulePort(diff *schema.ResourceDiff, v interface{}) error {
	_ = v

	// A value that is not known until apply reads back as the empty string
	// here. For protocol_type that would fall through to the port-independent
	// branch, and for dst_port it would look like a port the user never set,
	// so a perfectly good configuration that derives either one from another
	// resource would be rejected. There is nothing to check until both are
	// known.
	if !diff.NewValueKnown("protocol_type") || !diff.NewValueKnown("dst_port") {
		return nil
	}

	protocolType := diff.Get("protocol_type").(string)
	dstPort := diff.Get("dst_port").(string)
	// Spelled out from the list rather than written into the message, so the
	// two cannot drift apart when a protocol is added to it.
	portIndependent := strings.Join(secGroupPortIndependentProtocols, ", ")

	if isStringIn(protocolType, secGroupPortIndependentProtocols) {
		if dstPort != "" {
			return fmt.Errorf("%q must not be set when %q is one of %s", "dst_port", "protocol_type", portIndependent)
		}
		return nil
	}

	if dstPort == "" {
		return fmt.Errorf("%q must be set when %q is not one of %s", "dst_port", "protocol_type", portIndependent)
	}

	return nil
}

// secGroupRuleWaitForState waits for a rule to become readable. The API exposes
// no status for a rule, so this is an existence wait: statusPending means "not
// readable yet" and statusInitialized means "readable". Skipping it would let a
// read that lands too early clear the ID and drop the resource from state.
func secGroupRuleWaitForState(client *productClient, secGroupID, ruleID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			rule, err := client.describeSecGroupRuleById(secGroupID, ruleID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return rule, statusInitialized, nil
		},
	}
}
