package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudSubnetsRead,

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
				ForceNew: true,
			},

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"subnets": {
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

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"cidr_block": {
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

func dataSourceUCloudSubnetsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).vpcconn

	req := conn.NewDescribeSubnetRequest()

	if ids, ok := d.GetOk("ids"); ok {
		req.SubnetIds = schemaSetToStringSlice(ids)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	var allSubnets []vpc.VPCSubnetInfoSet
	var subnets []vpc.VPCSubnetInfoSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeSubnet(req)
		if err != nil {
			return fmt.Errorf("error on reading subnet list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allSubnets = append(allSubnets, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allSubnets {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			subnets = append(subnets, v)
		}
	} else {
		subnets = allSubnets
	}

	err := dataSourceUCloudSubnetsSave(d, subnets)
	if err != nil {
		return fmt.Errorf("error on reading subnet list, %s", err)
	}

	return nil
}

func dataSourceUCloudSubnetsSave(d *schema.ResourceData, subnets []vpc.VPCSubnetInfoSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, subnet := range subnets {
		ids = append(ids, string(subnet.SubnetId))

		data = append(data, map[string]interface{}{
			"id":          subnet.SubnetId,
			"name":        subnet.SubnetName,
			"create_time": timestampToString(subnet.CreateTime),
			"remark":      subnet.Remark,
			"tag":         subnet.Tag,
			"cidr_block":  fmt.Sprintf("%s/%s", subnet.Subnet, subnet.Netmask),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("subnets", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
