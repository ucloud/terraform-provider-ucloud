package ucloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func dataSourceUCloudIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudIAMPolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"System", "Custom"}, false),
			},

			"urn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"scope": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUCloudIAMPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	var owner string
	switch t := d.Get("type").(string); t {
	case "System":
		owner = "UCS"
	case "Custom":
		owner = "User"
	default:
		errors.New("type not supported")
	}
	policy, err := client.describeIAMPolicyByName(d.Get("name").(string), owner)
	if err != nil {
		return fmt.Errorf("error on reading policy: %s", err)
	}

	err = dataSourceUCloudIAMPolicySave(d, *policy)
	if err != nil {
		return fmt.Errorf("error on reading policy, %s", err)
	}

	return nil
}

func dataSourceUCloudIAMPolicySave(d *schema.ResourceData, policy iam.IAMPolicy) error {
	d.SetId(policy.PolicyURN)
	d.Set("urn", policy.PolicyURN)
	d.Set("comment", policy.Description)
	d.Set("create_time", timestampToString(policy.CreatedAt))
	switch policy.ScopeType {
	case "ScopeRequired":
		d.Set("scope", "Project")
	case "ScopeEmpty":
		d.Set("scope", "Account")
	case "ScopeUnrestricted":
		d.Set("scope", "Mixed")
	}
	d.Set("policy", policy.Document)
	return nil
}
