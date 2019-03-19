package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

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

						"expire_time": {
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
	var lbs []ulb.ULBSet
	var limit int = 100
	var totalCount int
	var offset int
	if ids, ok := d.GetOk("ids"); ok {
		idSet := schemaSetToStringSlice(ids)
		for _, v := range idSet {
			req := conn.NewDescribeULBRequest()
			req.ULBId = ucloud.String(v)
			resp, err := conn.DescribeULB(req)
			if err != nil {
				return fmt.Errorf("error on reading ulb list, %s", err)
			}

			lbs = append(lbs, resp.DataSet[0])
			totalCount++
		}
	} else {
		req := conn.NewDescribeULBRequest()
		for {
			req.Limit = ucloud.Int(limit)
			req.Offset = ucloud.Int(offset)
			resp, err := conn.DescribeULB(req)
			if err != nil {
				return fmt.Errorf("error on reading ulb list, %s", err)
			}

			if resp == nil || len(resp.DataSet) < 1 {
				break
			}

			lbs = append(lbs, resp.DataSet...)
			totalCount = totalCount + resp.TotalCount

			if len(resp.DataSet) < limit {
				break
			}

			offset = offset + limit
		}
	}

	d.Set("total_count", totalCount)
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

		data = append(data, map[string]interface{}{
			"id":          item.ULBId,
			"name":        item.Name,
			"tag":         item.Tag,
			"remark":      item.Remark,
			"vpc_id":      item.VPCId,
			"subnet_id":   item.SubnetId,
			"private_ip":  item.PrivateIP,
			"create_time": timestampToString(item.CreateTime),
			"expire_time": timestampToString(item.ExpireTime),
			"ip_set":      lbAddr,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("lbs", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
