package vpc

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudNatGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudNatGatewaysRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"nat_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":             {Type: schema.TypeString, Computed: true},
					"name":           {Type: schema.TypeString, Computed: true},
					"remark":         {Type: schema.TypeString, Computed: true},
					"tag":            {Type: schema.TypeString, Computed: true},
					"vpc_id":         {Type: schema.TypeString, Computed: true},
					"security_group": {Type: schema.TypeString, Computed: true},
					"create_time":    {Type: schema.TypeString, Computed: true},
					"subnet_ids": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"ip_set": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"ip":            {Type: schema.TypeString, Computed: true},
							"internet_type": {Type: schema.TypeString, Computed: true},
						}},
					},
				}},
			},
		},
	}
}

func dataSourceUCloudNatGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading nat gateway list, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewDescribeNATGWRequest()
	if ids, ok := d.GetOk("ids"); ok {
		req.NATGWIds = schemaSetToStringSlice(ids)
	}
	var allNatGateways []vpcapi.NatGatewayDataSet
	const limit = 100
	for offset := 0; ; offset += limit {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeNATGW(req)
		if err != nil {
			return fmt.Errorf("error on reading nat gateway list, %s", err)
		}
		if resp == nil || len(resp.DataSet) < 1 {
			break
		}
		allNatGateways = append(allNatGateways, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
	}
	natGateways := allNatGateways
	if value, ok := d.GetOk("name_regex"); ok {
		re := regexp.MustCompile(value.(string))
		natGateways = nil
		for _, item := range allNatGateways {
			if re.MatchString(item.NATGWName) {
				natGateways = append(natGateways, item)
			}
		}
	}
	if err := dataSourceUCloudNatGatewaysSave(d, natGateways); err != nil {
		return fmt.Errorf("error on reading natGateway list, %s", err)
	}
	return nil
}

func dataSourceUCloudNatGatewaysSave(d *schema.ResourceData, natGateways []vpcapi.NatGatewayDataSet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, natGateway := range natGateways {
		ids = append(ids, natGateway.NATGWId)
		ipSet := []map[string]interface{}{}
		for _, item := range natGateway.IPSet {
			ipSet = append(ipSet, map[string]interface{}{
				"ip": item.IPResInfo[0].EIP, "internet_type": item.IPResInfo[0].OperatorName,
			})
		}
		subnetIDs := []string{}
		for _, item := range natGateway.SubnetSet {
			subnetIDs = append(subnetIDs, item.SubnetworkId)
		}
		data = append(data, map[string]interface{}{
			"id": natGateway.NATGWId, "name": natGateway.NATGWName,
			"create_time": timestampToString(natGateway.CreateTime), "remark": natGateway.Remark,
			"tag": natGateway.Tag, "ip_set": ipSet, "subnet_ids": subnetIDs,
			"vpc_id": natGateway.VPCId, "security_group": natGateway.FirewallId,
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("nat_gateways", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
