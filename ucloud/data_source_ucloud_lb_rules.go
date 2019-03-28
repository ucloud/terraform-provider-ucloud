package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func dataSourceUCloudLBRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBRulesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"listener_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"lb_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudLBRulesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	var lbRules []ulb.ULBPolicySet
	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)
	vserverSet, err := client.describeVServerById(lbId, listenerId)
	if err != nil {
		return fmt.Errorf("error on reading lb rule list, %s", err)
	}

	if ids, ok := d.GetOk("ids"); ok {
		for _, v := range vserverSet.PolicySet {
			if !isStringIn(v.PolicyId, schemaSetToStringSlice(ids)) {
				continue
			}
			lbRules = append(lbRules, v)
		}
	} else {
		lbRules = vserverSet.PolicySet
	}

	err = dataSourceUCloudLBRulesSave(d, lbRules)
	if err != nil {
		return fmt.Errorf("error on reading lb rule list, %s", err)
	}

	return nil
}

func dataSourceUCloudLBRulesSave(d *schema.ResourceData, lbRules []ulb.ULBPolicySet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range lbRules {
		ids = append(ids, string(item.PolicyId))
		if item.Type == lbMatchTypePath {
			data = append(data, map[string]interface{}{
				"id":   item.PolicyId,
				"path": item.Match,
			})
		} else if item.Type == lbMatchTypeDomain {
			data = append(data, map[string]interface{}{
				"id":     item.PolicyId,
				"domain": item.Match,
			})
		}
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("lb_rules", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
