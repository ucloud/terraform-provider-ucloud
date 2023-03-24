package ucloud

import (
	"errors"
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/uads"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudAntiDDoSInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudAntiDDoSInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
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

			"instances": {
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

						"area": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"data_center": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"bandwidth": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"base_defence_value": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"max_defence_value": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"charge_type": {
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

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudAntiDDoSInstancesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uadsconn
	var allInstances []uads.ServiceInfo
	var instances []uads.ServiceInfo

	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeNapServiceInfoRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		req.NapType = ucloud.Int(1)
		resp, err := conn.DescribeNapServiceInfo(req)
		if err != nil {
			return fmt.Errorf("error on reading anti ddos instance list, %s", err)
		}

		if resp == nil || len(resp.ServiceInfo) < 1 {
			break
		}

		allInstances = append(allInstances, resp.ServiceInfo...)

		if len(resp.ServiceInfo) < limit {
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
		for _, v := range allInstances {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.ResourceId, schemaSetToStringSlice(ids)) {
				continue
			}
			instances = append(instances, v)
		}
	} else {
		instances = allInstances
	}

	err := dataSourceUCloudAntiDDoSInstancesSave(d, instances)
	if err != nil {
		return fmt.Errorf("error on reading disk list, %s", err)
	}

	return nil
}

func dataSourceUCloudAntiDDoSInstancesSave(d *schema.ResourceData, instances []uads.ServiceInfo) error {
	var ids []string
	var data []map[string]interface{}

	for _, item := range instances {
		ids = append(ids, item.ResourceId)
		if len(item.EngineRoom) < 1 {
			return errors.New("fail to get data_center")
		}
		if len(item.DefenceDDosBaseFlowArr) < 1 {
			return errors.New("fail to get base_defence_value")
		}
		if len(item.DefenceDDosMaxFlowArr) < 1 {
			return errors.New("fail to get max_defence_value")
		}
		data = append(data, map[string]interface{}{
			"id":                 item.ResourceId,
			"charge_type":        upperCamelCvt.convert(item.ChargeType),
			"name":               item.Name,
			"area":               item.AreaLine,
			"data_center":        item.EngineRoom[0],
			"bandwidth":          item.SrcBandwidth,
			"base_defence_value": item.DefenceDDosBaseFlowArr[0],
			"max_defence_value":  item.DefenceDDosMaxFlowArr[0],
			"create_time":        timestampToString(item.CreateTime),
			"expire_time":        timestampToString(item.ExpiredTime),
			"status":             item.DefenceStatus,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("instances", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
