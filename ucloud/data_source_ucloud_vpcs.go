package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudVPCs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudVPCsRead,

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

			"vpcs": {
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

						"cidr_blocks": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"update_time": {
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

func dataSourceUCloudVPCsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).vpcconn
	var allVPCs []vpc.VPCInfo
	var vpcs []vpc.VPCInfo
	var limit int = 100
	var offset int

	req := conn.NewDescribeVPCRequest()
	if ids, ok := d.GetOk("ids"); ok {
		req.VPCIds = schemaSetToStringSlice(ids)
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	for {
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

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allVPCs {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			vpcs = append(vpcs, v)
		}
	} else {
		vpcs = allVPCs
	}

	err := dataSourceUCloudVPCsSave(d, vpcs)
	if err != nil {
		return fmt.Errorf("error on reading vpc list, %s", err)
	}

	return nil
}

func dataSourceUCloudVPCsSave(d *schema.ResourceData, vpcs []vpc.VPCInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, vpc := range vpcs {
		ids = append(ids, string(vpc.VPCId))

		data = append(data, map[string]interface{}{
			"id":          vpc.VPCId,
			"name":        vpc.Name,
			"create_time": timestampToString(vpc.CreateTime),
			"update_time": timestampToString(vpc.UpdateTime),
			"tag":         vpc.Tag,
			"cidr_blocks": vpc.Network,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("vpcs", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
