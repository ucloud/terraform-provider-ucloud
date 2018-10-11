package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudEips() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudEipsRead,

		Schema: map[string]*schema.Schema{
			"ids": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"eips": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_set": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},

									"internet_type": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"bandwidth": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"internet_charge_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"internet_charge_mode": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"create_time": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"expire_time": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudEipsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).unetconn

	req := conn.NewDescribeEIPRequest()

	if ids, ok := d.GetOk("ids"); ok && len(ids.([]interface{})) > 0 {
		req.EIPIds = ifaceToStringSlice(ids)
	}

	var eips []unet.UnetEIPSet
	var limit int = 100
	var totalCount int
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeEIP(req)
		if err != nil {
			return fmt.Errorf("error in read eip list, %s", err)
		}

		if resp == nil || len(resp.EIPSet) < 1 {
			break
		}

		eips = append(eips, resp.EIPSet...)

		totalCount = totalCount + resp.TotalCount

		if len(resp.EIPSet) < limit {
			break
		}

		offset = offset + limit
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudEipsSave(d, eips)
	if err != nil {
		return fmt.Errorf("error in read eip list, %s", err)
	}

	return nil
}

func dataSourceUCloudEipsSave(d *schema.ResourceData, eips []unet.UnetEIPSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range eips {
		ids = append(ids, string(item.EIPId))

		eipAddr := []map[string]string{}
		for _, addr := range item.EIPAddr {
			eipAddr = append(eipAddr, map[string]string{
				"ip":            addr.IP,
				"internet_type": addr.OperatorName,
			})
		}

		data = append(data, map[string]interface{}{
			"bandwidth":            item.Bandwidth,
			"internet_charge_type": item.ChargeType,
			"internet_charge_mode": item.PayMode,
			"name":                 item.Name,
			"remark":               item.Remark,
			"tag":                  item.Tag,
			"status":               item.Status,
			"create_time":          timestampToString(item.CreateTime),
			"expire_time":          timestampToString(item.ExpireTime),
			"ip_set":               eipAddr,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("eips", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
