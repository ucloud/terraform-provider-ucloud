package ulb

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBRuleCreate,
		Update: resourceUCloudLBRuleUpdate,
		Read:   resourceUCloudLBRuleRead,
		Delete: resourceUCloudLBRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.All(
			customizeDiffLBRuleDomainWithPath,
		),
		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backend_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
				Set:      schema.HashString,
			},
			"domain": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"path"},
			},
			"path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"domain"},
			},
		},
	}
}

func resourceUCloudLBRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb rule, %s", err)
	}
	conn := client.ulbconn
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	listenerSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return fmt.Errorf("error on reading lb listener when creating lb rule, %s", err)
	}
	protocol := listenerSet.Protocol
	if protocol != "HTTP" && protocol != "HTTPS" {
		return fmt.Errorf("the lb rule can only be define while the protocol of lb listener is one of http and https, got %s", upperCvt.convert(protocol))
	}
	req := conn.NewCreatePolicyRequest()
	req.ULBId = ucloud.String(lbID)
	req.VServerId = ucloud.String(listenerID)
	req.BackendId = schemaSetToStringSlice(d.Get("backend_ids"))
	if value, ok := d.GetOk("domain"); ok {
		req.Type = ucloud.String(lbMatchTypeDomain)
		req.Match = ucloud.String(value.(string))
	} else if value, ok := d.GetOk("path"); ok {
		req.Type = ucloud.String(lbMatchTypePath)
		req.Match = ucloud.String(value.(string))
	} else {
		return fmt.Errorf("error on creating lb rule, shoule set one of domain and path")
	}
	resp, err := conn.CreatePolicy(req)
	if err != nil {
		return fmt.Errorf("error on creating lb rule, %s", err)
	}
	d.SetId(resp.PolicyId)
	if _, err = lbRuleWaitForState(client, lbID, listenerID, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for lb rule %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudLBRuleRead(d, meta)
}

func resourceUCloudLBRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating lb rule, %s", err)
	}
	conn := client.ulbconn
	d.Partial(true)
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	req := conn.NewUpdatePolicyRequest()
	req.ULBId = ucloud.String(lbID)
	req.VServerId = ucloud.String(listenerID)
	req.BackendId = schemaSetToStringSlice(d.Get("backend_ids"))
	req.PolicyId = ucloud.String(d.Id())
	isChanged := false
	if d.HasChange("domain") && !d.IsNewResource() {
		isChanged = true
		req.Type = ucloud.String(lbMatchTypeDomain)
		req.Match = ucloud.String(d.Get("domain").(string))
	}
	if d.HasChange("path") && !d.IsNewResource() {
		isChanged = true
		req.Type = ucloud.String(lbMatchTypePath)
		req.Match = ucloud.String(d.Get("path").(string))
	}
	if isChanged {
		if _, err := conn.UpdatePolicy(req); err != nil {
			return fmt.Errorf("error on %s to lb rule %q, %s", "UpdatePolicy", d.Id(), err)
		}
		d.SetPartial("domain")
		d.SetPartial("path")
		if _, err := lbRuleWaitForState(client, lbID, listenerID, d.Id()).WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for %s complete to lb rule %q, %s", "UpdatePolicy", d.Id(), err)
		}
	}
	d.Partial(false)
	return resourceUCloudLBRuleRead(d, meta)
}

func resourceUCloudLBRuleRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb rule, %s", err)
	}
	var lbID, listenerID string
	var policySet *ulb.ULBPolicySet
	if value, ok := d.GetOk("load_balancer_id"); ok {
		listenerID = d.Get("listener_id").(string)
		policySet, err = client.describePolicyById(value.(string), listenerID, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb rule %q, %s", d.Id(), err)
		}
		_ = d.Set("load_balancer_id", value)
		_ = d.Set("listener_id", listenerID)
	} else {
		policySet, lbID, listenerID, err = client.describePolicyByOneId(d.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb rule %q, %s", d.Id(), err)
		}
		_ = d.Set("load_balancer_id", lbID)
		_ = d.Set("listener_id", listenerID)
	}
	if policySet.Type == lbMatchTypePath {
		_ = d.Set("path", policySet.Match)
	}
	if policySet.Type == lbMatchTypeDomain {
		_ = d.Set("domain", policySet.Match)
	}
	backendIDs := []string{}
	for _, item := range policySet.BackendSet {
		backendIDs = append(backendIDs, item.BackendId)
	}
	_ = d.Set("backend_ids", backendIDs)
	return nil
}

func resourceUCloudLBRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb rule, %s", err)
	}
	conn := client.ulbconn
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	req := conn.NewDeletePolicyRequest()
	req.VServerId = ucloud.String(listenerID)
	req.PolicyId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeletePolicy(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb rule %q, %s", d.Id(), err))
		}
		if _, err := client.describePolicyById(lbID, listenerID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb rule when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified lb rule %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbRuleWaitForState(client *productClient, lbID, listenerID, policyID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			policySet, err := client.describePolicyById(lbID, listenerID, policyID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return policySet, statusInitialized, nil
		},
	}
}

func customizeDiffLBRuleDomainWithPath(diff *schema.ResourceDiff, _ interface{}) error {
	_, pathOK := diff.GetOk("path")
	_, domainOK := diff.GetOk("domain")
	if !pathOK && !domainOK {
		return fmt.Errorf("should set one of domain and path")
	}
	return nil
}
