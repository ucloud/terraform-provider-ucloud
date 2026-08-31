package iam

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMPolicyCreate,
		Update: resourceUCloudIAMPolicyUpdate,
		Read:   resourceUCloudIAMPolicyRead,
		Delete: resourceUCloudIAMPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"Project",
					"Account",
				}, false),
				Required: true,
				ForceNew: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"urn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIAMPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewCreateIAMPolicyRequest()
	name := d.Get("name").(string)
	req.PolicyName = ucloud.String(name)
	req.Description = ucloud.String(d.Get("comment").(string))
	req.Document = ucloud.String(d.Get("policy").(string))
	switch scope := d.Get("scope").(string); scope {
	case "Project":
		req.ScopeType = ucloud.String("ScopeRequired")
	case "Account":
		req.ScopeType = ucloud.String("ScopeEmpty")
	}
	if _, err := client.iamconn.CreateIAMPolicy(req); err != nil {
		return fmt.Errorf("error on creating iam policy, %s", err)
	}
	d.SetId(name)
	return resourceUCloudIAMPolicyRead(d, meta)
}

func resourceUCloudIAMPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	d.Partial(true)
	if d.HasChanges("comment") {
		req := client.iamconn.NewUpdateIAMPolicyNameRequest()
		req.PolicyName = ucloud.String(d.Get("name").(string))
		req.PolicyURN = ucloud.String(d.Get("urn").(string))
		req.Description = ucloud.String(d.Get("comment").(string))
		if _, err := client.iamconn.UpdateIAMPolicyName(req); err != nil {
			return fmt.Errorf("error on %s to update policy, %s", d.Id(), err)
		}
		d.SetPartial("comment")
	}
	if d.HasChanges("policy") {
		req := client.iamconn.NewUpdateIAMPolicyRequest()
		req.PolicyURN = ucloud.String(d.Get("urn").(string))
		req.Document = ucloud.String(d.Get("policy").(string))
		if _, err := client.iamconn.UpdateIAMPolicy(req); err != nil {
			return fmt.Errorf("error on %s to update policy, %s", d.Id(), err)
		}
		d.SetPartial("policy")
	}
	d.Partial(false)
	return resourceUCloudIAMPolicyRead(d, meta)
}

func resourceUCloudIAMPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	var policy *iam.IAMPolicy
	if urn, ok := d.GetOk("urn"); ok {
		policy, err = client.describeIAMPolicyByURN(urn.(string))
	} else {
		policy, err = client.describeIAMPolicyByName(d.Id(), "User")
	}
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading policy %q, %s", d.Id(), err)
	}

	switch policy.ScopeType {
	case "ScopeRequired":
		_ = d.Set("scope", "Project")
	case "ScopeEmpty":
		_ = d.Set("scope", "Account")
	}
	_ = d.Set("name", policy.PolicyName)
	_ = d.Set("comment", policy.Description)
	_ = d.Set("policy", policy.Document)
	_ = d.Set("urn", policy.PolicyURN)
	_ = d.Set("create_time", timestampToString(policy.CreatedAt))
	return nil
}

func resourceUCloudIAMPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewDeleteIAMPolicyRequest()
	req.PolicyURN = ucloud.String(d.Get("urn").(string))
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.iamconn.DeleteIAMPolicy(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting policy %q, %s", d.Id(), err))
		}

		_, err := client.describeIAMPolicyByName(d.Id(), "User")
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading policy when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified policy %q has not been deleted due to unknown error", d.Id()))
	})
}
