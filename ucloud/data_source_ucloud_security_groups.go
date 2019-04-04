package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudSecurityGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudSecurityGroupsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"rules": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port_range": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"cidr_block": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"policy": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"priority": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudSecurityGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).unetconn
	var allSecurityGroups []unet.FirewallDataSet
	var securityGroups []unet.FirewallDataSet
	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeFirewallRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeFirewall(req)
		if err != nil {
			return fmt.Errorf("error on reading security group list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allSecurityGroups = append(allSecurityGroups, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if idsOk || nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allSecurityGroups {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.FWId, schemaSetToStringSlice(ids)) {
				continue
			}
			securityGroups = append(securityGroups, v)
		}
	} else {
		securityGroups = allSecurityGroups
	}

	err := dataSourceUCloudSecurityGroupsSave(d, securityGroups)
	if err != nil {
		return fmt.Errorf("error on reading security group list, %s", err)
	}

	return nil
}

func dataSourceUCloudSecurityGroupsSave(d *schema.ResourceData, securityGroups []unet.FirewallDataSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range securityGroups {
		ids = append(ids, string(item.FWId))

		rules := []map[string]interface{}{}
		for _, v := range item.Rule {
			rules = append(rules, map[string]interface{}{
				"port_range": v.DstPort,
				"protocol":   upperCvt.convert(v.ProtocolType),
				"cidr_block": v.SrcIP,
				"policy":     upperCvt.convert(v.RuleAction),
				"priority":   upperCvt.convert(v.Priority),
			})
		}

		data = append(data, map[string]interface{}{
			"id":          item.FWId,
			"remark":      item.Remark,
			"name":        item.Name,
			"tag":         item.Tag,
			"rules":       rules,
			"create_time": timestampToString(item.CreateTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("security_groups", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
