package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/customdiff"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	listenerSet, err := client.describeVServerById(lbId, listenerId)
	if err != nil {
		return fmt.Errorf("error on reading lb listener when creating lb rule, %s", err)
	}
	protocol := listenerSet.Protocol
	if protocol != "HTTP" && protocol != "HTTPS" {
		return fmt.Errorf("the lb rule can only be define while the protocol of lb listener is one of http and https, got %s", upperCvt.convert(protocol))
	}

	req := conn.NewCreatePolicyRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)
	req.BackendId = schemaSetToStringSlice(d.Get("backend_ids"))

	if val, ok := d.GetOk("domain"); ok {
		req.Type = ucloud.String("Domain")
		req.Match = ucloud.String(val.(string))
	} else if val, ok := d.GetOk("path"); ok {
		req.Type = ucloud.String("Path")
		req.Match = ucloud.String(val.(string))
	} else {
		return fmt.Errorf("error on creating lb rule, shoule set one of domain and path")
	}

	resp, err := conn.CreatePolicy(req)

	if err != nil {
		return fmt.Errorf("error on creating lb rule, %s", err)
	}

	d.SetId(resp.PolicyId)

	// after create lb rule, we need to wait it initialized
	stateConf := lbRuleWaitForState(client, lbId, listenerId, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for lb rule %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudLBRuleRead(d, meta)
}

func resourceUCloudLBRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	d.Partial(true)

	isChanged := false
	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	req := conn.NewUpdatePolicyRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)
	req.BackendId = schemaSetToStringSlice(d.Get("backend_ids"))
	req.PolicyId = ucloud.String(d.Id())

	if d.HasChange("domain") && !d.IsNewResource() {
		isChanged = true
		req.Type = ucloud.String("Domain")
		req.Match = ucloud.String(d.Get("domain").(string))
	}

	if d.HasChange("path") && !d.IsNewResource() {
		isChanged = true
		req.Type = ucloud.String("Path")
		req.Match = ucloud.String(d.Get("path").(string))
	}

	if isChanged {
		_, err := conn.UpdatePolicy(req)

		if err != nil {
			return fmt.Errorf("error on %s to lb rule %q, %s", "UpdatePolicy", d.Id(), err)
		}

		d.SetPartial("domain")
		d.SetPartial("path")

		// after update lb rule, we need to wait it completed
		stateConf := lbRuleWaitForState(client, lbId, listenerId, d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error on waiting for %s complete to lb rule %q, %s", "UpdatePolicy", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudLBRuleRead(d, meta)
}

func resourceUCloudLBRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	var err error
	var lbId string
	var listenerId string
	var policySet *ulb.ULBPolicySet
	if v, ok := d.GetOk("load_balancer_id"); ok {
		listenerId = d.Get("listener_id").(string)
		policySet, err = client.describePolicyById(v.(string), listenerId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb rule %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", v)
		d.Set("listener_id", listenerId)
	} else {
		policySet, lbId, listenerId, err = client.describePolicyByOneId(d.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb rule %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", lbId)
		d.Set("listener_id", listenerId)
	}

	if policySet.Type == lbMatchTypePath {
		d.Set("path", policySet.Match)
	}

	if policySet.Type == lbMatchTypeDomain {
		d.Set("domain", policySet.Match)
	}

	backendIds := []string{}
	for _, item := range policySet.BackendSet {
		backendIds = append(backendIds, item.BackendId)
	}

	d.Set("backend_ids", backendIds)

	return nil
}

func resourceUCloudLBRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	req := conn.NewDeletePolicyRequest()
	req.VServerId = ucloud.String(listenerId)
	req.PolicyId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeletePolicy(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb rule %q, %s", d.Id(), err))
		}

		_, err := client.describePolicyById(lbId, listenerId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb rule when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb rule %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbRuleWaitForState(client *UCloudClient, lbId, listenerId, policyId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			policySet, err := client.describePolicyById(lbId, listenerId, policyId)
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

func customizeDiffLBRuleDomainWithPath(diff *schema.ResourceDiff, v interface{}) error {
	_, pOk := diff.GetOk("path")
	_, dOk := diff.GetOk("domain")
	if !pOk && !dOk {
		return fmt.Errorf("should set one of domain and path")
	}
	return nil
}
