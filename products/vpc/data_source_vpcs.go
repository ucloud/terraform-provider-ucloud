package vpc

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudVPCs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudVPCsRead,
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
			"tag":         {Type: schema.TypeString, Optional: true},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"vpcs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":          {Type: schema.TypeString, Computed: true},
					"name":        {Type: schema.TypeString, Computed: true},
					"cidr_blocks": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
					"tag":         {Type: schema.TypeString, Computed: true},
					"update_time": {Type: schema.TypeString, Computed: true},
					"create_time": {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudVPCsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading vpc list, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewDescribeVPCRequest()
	if ids, ok := d.GetOk("ids"); ok {
		req.VPCIds = schemaSetToStringSlice(ids)
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	}
	var allVPCs []vpcapi.VPCInfo
	const limit = 100
	for offset := 0; ; offset += limit {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeVPC(req)
		if err != nil {
			return fmt.Errorf("error on reading vpc list, %s", err)
		}
		if resp == nil || len(resp.DataSet) < 1 {
			break
		}
		allVPCs = append(allVPCs, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
	}
	vpcs := allVPCs
	if value, ok := d.GetOk("name_regex"); ok {
		re := regexp.MustCompile(value.(string))
		vpcs = nil
		for _, item := range allVPCs {
			if re.MatchString(item.Name) {
				vpcs = append(vpcs, item)
			}
		}
	}
	if err := dataSourceUCloudVPCsSave(d, vpcs); err != nil {
		return fmt.Errorf("error on reading vpc list, %s", err)
	}
	return nil
}

func dataSourceUCloudVPCsSave(d *schema.ResourceData, vpcs []vpcapi.VPCInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range vpcs {
		ids = append(ids, item.VPCId)
		data = append(data, map[string]interface{}{
			"id": item.VPCId, "name": item.Name,
			"create_time": timestampToString(item.CreateTime), "update_time": timestampToString(item.UpdateTime),
			"tag": item.Tag, "cidr_blocks": item.Network,
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("vpcs", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
