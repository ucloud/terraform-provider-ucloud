package ulb

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
				Set:      schema.HashString,
				Computed: true,
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
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"internal":    {Type: schema.TypeBool, Computed: true},
						"tag":         {Type: schema.TypeString, Computed: true},
						"remark":      {Type: schema.TypeString, Computed: true},
						"vpc_id":      {Type: schema.TypeString, Computed: true},
						"subnet_id":   {Type: schema.TypeString, Computed: true},
						"private_ip":  {Type: schema.TypeString, Computed: true},
						"create_time": {Type: schema.TypeString, Computed: true},
						"ip_set": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"ip":            {Type: schema.TypeString, Computed: true},
								"internet_type": {Type: schema.TypeString, Computed: true},
							}},
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudLBsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading ulb list, %s", err)
	}
	conn := client.ulbconn
	var allLBs []ulb.ULBSet
	const limit = 100
	for offset := 0; ; offset += limit {
		req := conn.NewDescribeULBRequest()
		if value, ok := d.GetOk("vpc_id"); ok {
			req.VPCId = ucloud.String(value.(string))
		}
		if value, ok := d.GetOk("subnet_id"); ok {
			req.SubnetId = ucloud.String(value.(string))
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
		allLBs = append(allLBs, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
	}

	ids, idsOK := d.GetOk("ids")
	nameRegex, regexOK := d.GetOk("name_regex")
	var lbs []ulb.ULBSet
	if idsOK || regexOK {
		var re *regexp.Regexp
		if regexOK {
			re = regexp.MustCompile(nameRegex.(string))
		}
		for _, item := range allLBs {
			if re != nil && !re.MatchString(item.Name) {
				continue
			}
			if idsOK && !isStringIn(item.ULBId, schemaSetToStringSlice(ids)) {
				continue
			}
			lbs = append(lbs, item)
		}
	} else {
		lbs = allLBs
	}
	if err := dataSourceUCloudLBsSave(d, lbs); err != nil {
		return fmt.Errorf("error on reading ulb list, %s", err)
	}
	return nil
}

func dataSourceUCloudLBsSave(d *schema.ResourceData, lbs []ulb.ULBSet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range lbs {
		ids = append(ids, item.ULBId)
		ipSet := []map[string]string{}
		for _, addr := range item.IPSet {
			ipSet = append(ipSet, map[string]string{"ip": addr.EIP, "internet_type": addr.OperatorName})
		}
		var internal bool
		if item.ULBType == "InnerMode" {
			internal = true
		}
		data = append(data, map[string]interface{}{
			"id": item.ULBId, "name": item.Name, "internal": internal,
			"tag": item.Tag, "remark": item.Remark, "vpc_id": item.VPCId,
			"subnet_id": item.SubnetId, "private_ip": item.PrivateIP,
			"create_time": timestampToString(item.CreateTime), "ip_set": ipSet,
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("lbs", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
