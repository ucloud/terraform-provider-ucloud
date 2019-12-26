package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudLBs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBsRead,

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

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"lbs": {
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

						"internal": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"private_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ip_set": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"internet_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
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

func dataSourceUCloudLBsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn
	var allLbs []ulb.ULBSet
	var lbs []ulb.ULBSet
	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeULBRequest()
		if v, ok := d.GetOk("vpc_id"); ok {
			req.VPCId = ucloud.String(v.(string))
		}

		if v, ok := d.GetOk("subnet_id"); ok {
			req.SubnetId = ucloud.String(v.(string))
		}

		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeULB(req)
		if err != nil {
			return fmt.Errorf("error on reading ulb list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allLbs = append(allLbs, resp.DataSet...)

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
		for _, v := range allLbs {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.ULBId, schemaSetToStringSlice(ids)) {
				continue
			}
			lbs = append(lbs, v)
		}
	} else {
		lbs = allLbs
	}

	err := dataSourceUCloudLBsSave(d, lbs)
	if err != nil {
		return fmt.Errorf("error on reading ulb list, %s", err)
	}

	return nil
}

func dataSourceUCloudLBsSave(d *schema.ResourceData, lbs []ulb.ULBSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range lbs {
		ids = append(ids, string(item.ULBId))

		lbAddr := []map[string]string{}
		for _, addr := range item.IPSet {
			lbAddr = append(lbAddr, map[string]string{
				"ip":            addr.EIP,
				"internet_type": addr.OperatorName,
			})
		}
		var internal bool
		if item.ULBType == "OuterMode" {
			internal = false
		} else if item.ULBType == "InnerMode" {
			internal = true
		}

		data = append(data, map[string]interface{}{
			"id":          item.ULBId,
			"name":        item.Name,
			"internal":    internal,
			"tag":         item.Tag,
			"remark":      item.Remark,
			"vpc_id":      item.VPCId,
			"subnet_id":   item.SubnetId,
			"private_ip":  item.PrivateIP,
			"create_time": timestampToString(item.CreateTime),
			"ip_set":      lbAddr,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("lbs", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
