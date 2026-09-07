package vpc

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func dataSourceUCloudSecGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudSecGroupsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"vpc_id":      {Type: schema.TypeString, Optional: true},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"sec_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":     {Type: schema.TypeString, Computed: true},
					"name":   {Type: schema.TypeString, Computed: true},
					"vpc_id": {Type: schema.TypeString, Computed: true},
					"type":   {Type: schema.TypeString, Computed: true},
					"rules": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"direction":     {Type: schema.TypeString, Computed: true},
							"protocol_type": {Type: schema.TypeString, Computed: true},
							"dst_port":      {Type: schema.TypeString, Computed: true},
							"ip_range":      {Type: schema.TypeString, Computed: true},
							"rule_action":   {Type: schema.TypeString, Computed: true},
							"priority":      {Type: schema.TypeInt, Computed: true},
							"remark":        {Type: schema.TypeString, Computed: true},
							"rule_id":       {Type: schema.TypeString, Computed: true},
						}},
					},
					"tag":         {Type: schema.TypeString, Computed: true},
					"remark":      {Type: schema.TypeString, Computed: true},
					"create_time": {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudSecGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading sec group list, %s", err)
	}
	vpcID := ""
	if value, ok := d.GetOk("vpc_id"); ok {
		vpcID = value.(string)
	}
	allSecGroups, err := client.describeSecGroupsByVPCId(vpcID)
	if err != nil {
		return fmt.Errorf("error on reading sec group list, %s", err)
	}
	ids, idsOK := d.GetOk("ids")
	nameRegex, regexOK := d.GetOk("name_regex")
	var secGroups []vpcapi.SecGroupInfo
	if idsOK || regexOK {
		var re *regexp.Regexp
		if regexOK {
			re = regexp.MustCompile(nameRegex.(string))
		}
		for _, item := range allSecGroups {
			if re != nil && !re.MatchString(item.Name) {
				continue
			}
			if idsOK && !isStringIn(item.SecGroupId, schemaSetToStringSlice(ids)) {
				continue
			}
			secGroups = append(secGroups, item)
		}
	} else {
		secGroups = allSecGroups
	}
	if err := dataSourceUCloudSecGroupsSave(d, secGroups); err != nil {
		return fmt.Errorf("error on reading sec group list, %s", err)
	}
	return nil
}

func dataSourceUCloudSecGroupsSave(d *schema.ResourceData, secGroups []vpcapi.SecGroupInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range secGroups {
		ids = append(ids, item.SecGroupId)
		rules := []map[string]interface{}{}
		for _, rule := range item.Rule {
			rules = append(rules, map[string]interface{}{
				"direction": rule.Direction, "protocol_type": rule.ProtocolType,
				"dst_port": rule.DstPort, "ip_range": rule.IPRange, "rule_action": rule.RuleAction,
				"priority": rule.Priority, "remark": rule.Remark, "rule_id": rule.RuleId,
			})
		}
		data = append(data, map[string]interface{}{
			"id": item.SecGroupId, "name": item.Name, "vpc_id": item.VPCId, "type": item.Type,
			"tag": item.Tag, "remark": item.Remark, "rules": rules,
			"create_time": timestampToString(item.CreateTime),
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("sec_groups", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
