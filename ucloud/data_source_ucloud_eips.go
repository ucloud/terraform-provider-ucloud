package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudEips() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudEipsRead,

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

			"eips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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

						"bandwidth": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"charge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"charge_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
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

func dataSourceUCloudEipsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).unetconn

	req := conn.NewDescribeEIPRequest()

	if ids, ok := d.GetOk("ids"); ok {
		req.EIPIds = schemaSetToStringSlice(ids)
	}

	var allEips []unet.UnetEIPSet
	var eips []unet.UnetEIPSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeEIP(req)
		if err != nil {
			return fmt.Errorf("error on reading eip list, %s", err)
		}

		if resp == nil || len(resp.EIPSet) < 1 {
			break
		}

		allEips = append(allEips, resp.EIPSet...)

		if len(resp.EIPSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allEips {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}
			eips = append(eips, v)
		}
	} else {
		eips = allEips
	}

	err := dataSourceUCloudEipsSave(d, eips)
	if err != nil {
		return fmt.Errorf("error on reading eip list, %s", err)
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
			"bandwidth":   item.Bandwidth,
			"charge_type": upperCamelCvt.convert(item.ChargeType),
			"charge_mode": upperCamelCvt.convert(item.PayMode),
			"name":        item.Name,
			"remark":      item.Remark,
			"tag":         item.Tag,
			"status":      item.Status,
			"create_time": timestampToString(item.CreateTime),
			"expire_time": timestampToString(item.ExpireTime),
			"ip_set":      eipAddr,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("eips", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
