package vpc

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudSubnetsRead,
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
			"vpc_id":      {Type: schema.TypeString, Optional: true, ForceNew: true},
			"tag":         {Type: schema.TypeString, Optional: true},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":          {Type: schema.TypeString, Computed: true},
					"name":        {Type: schema.TypeString, Computed: true},
					"remark":      {Type: schema.TypeString, Computed: true},
					"tag":         {Type: schema.TypeString, Computed: true},
					"cidr_block":  {Type: schema.TypeString, Computed: true},
					"create_time": {Type: schema.TypeString, Computed: true},
					"vpc_id":      {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudSubnetsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading subnet list, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewDescribeSubnetRequest()
	if ids, ok := d.GetOk("ids"); ok {
		req.SubnetIds = schemaSetToStringSlice(ids)
	}
	if value, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	}
	var allSubnets []vpcapi.SubnetInfo
	const limit = 100
	for offset := 0; ; offset += limit {
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
	}
	subnets := allSubnets
	if value, ok := d.GetOk("name_regex"); ok {
		re := regexp.MustCompile(value.(string))
		subnets = nil
		for _, item := range allSubnets {
			if re.MatchString(item.SubnetName) {
				subnets = append(subnets, item)
			}
		}
	}
	if err := dataSourceUCloudSubnetsSave(d, subnets); err != nil {
		return fmt.Errorf("error on reading subnet list, %s", err)
	}
	return nil
}

func dataSourceUCloudSubnetsSave(d *schema.ResourceData, subnets []vpcapi.SubnetInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range subnets {
		ids = append(ids, item.SubnetId)
		data = append(data, map[string]interface{}{
			"id": item.SubnetId, "name": item.SubnetName, "create_time": timestampToString(item.CreateTime),
			"remark": item.Remark, "tag": item.Tag,
			"cidr_block": fmt.Sprintf("%s/%s", item.Subnet, item.Netmask), "vpc_id": item.VPCId,
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("subnets", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		if err := writeToFile(outputFile.(string), data); err != nil {
			return err
		}
	}
	return nil
}
