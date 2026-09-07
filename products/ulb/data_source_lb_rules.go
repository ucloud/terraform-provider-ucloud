package ulb

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func dataSourceUCloudLBRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBRulesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"load_balancer_id": {Type: schema.TypeString, Required: true},
			"listener_id":      {Type: schema.TypeString, Required: true},
			"output_file":      {Type: schema.TypeString, Optional: true},
			"total_count":      {Type: schema.TypeInt, Computed: true},
			"lb_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":     {Type: schema.TypeString, Computed: true},
					"domain": {Type: schema.TypeString, Computed: true},
					"path":   {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudLBRulesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb rule list, %s", err)
	}
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return fmt.Errorf("error on reading lb rule list, %s", err)
	}
	var rules []ulb.ULBPolicySet
	if ids, ok := d.GetOk("ids"); ok {
		for _, item := range vserverSet.PolicySet {
			if isStringIn(item.PolicyId, schemaSetToStringSlice(ids)) {
				rules = append(rules, item)
			}
		}
	} else {
		rules = vserverSet.PolicySet
	}
	if err := dataSourceUCloudLBRulesSave(d, rules); err != nil {
		return fmt.Errorf("error on reading lb rule list, %s", err)
	}
	return nil
}

func dataSourceUCloudLBRulesSave(d *schema.ResourceData, rules []ulb.ULBPolicySet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range rules {
		ids = append(ids, item.PolicyId)
		switch item.Type {
		case lbMatchTypePath:
			data = append(data, map[string]interface{}{"id": item.PolicyId, "path": item.Match})
		case lbMatchTypeDomain:
			data = append(data, map[string]interface{}{"id": item.PolicyId, "domain": item.Match})
		}
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("lb_rules", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
