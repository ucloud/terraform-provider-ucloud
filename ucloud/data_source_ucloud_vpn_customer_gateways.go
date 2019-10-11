package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dateSourceUCloudVPNCustomerGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudVPNCustomerGatewaysRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
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

			"vpn_customer_gateways": {
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

						"ip_address": {
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

func dataSourceUCloudVPNCustomerGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ipsecvpnClient

	req := conn.NewDescribeRemoteVPNGatewayRequest()

	if v, ok := d.GetOk("ids"); ok {
		req.RemoteVPNGatewayIds = schemaSetToStringSlice(v)
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	var allVPNCusGateways []ipsecvpn.RemoteVPNGatewayDataSet
	var vpnCusGateways []ipsecvpn.RemoteVPNGatewayDataSet
	var limit = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeRemoteVPNGateway(req)
		if err != nil {
			return fmt.Errorf("error on reading vpn customer gateway list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allVPNCusGateways = append(allVPNCusGateways, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allVPNCusGateways {
			if r != nil && !r.MatchString(v.RemoteVPNGatewayName) {
				continue
			}

			vpnCusGateways = append(vpnCusGateways, v)
		}
	} else {
		vpnCusGateways = allVPNCusGateways
	}

	err := dataSourceUCloudVPNCustomerGatewaysSave(d, vpnCusGateways)
	if err != nil {
		return fmt.Errorf("error on reading vpnCusGateway list, %s", err)
	}

	return nil
}

func dataSourceUCloudVPNCustomerGatewaysSave(d *schema.ResourceData, vpnCusGateways []ipsecvpn.RemoteVPNGatewayDataSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, vpnCusGateway := range vpnCusGateways {
		ids = append(ids, vpnCusGateway.RemoteVPNGatewayId)
		data = append(data, map[string]interface{}{
			"id":          vpnCusGateway.RemoteVPNGatewayId,
			"name":        vpnCusGateway.RemoteVPNGatewayName,
			"ip_address":  vpnCusGateway.RemoteVPNGatewayAddr,
			"remark":      vpnCusGateway.Remark,
			"tag":         vpnCusGateway.Tag,
			"create_time": timestampToString(vpnCusGateway.CreateTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("vpn_customer_gateways", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
