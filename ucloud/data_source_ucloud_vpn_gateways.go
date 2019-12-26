package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dateSourceUCloudVPNGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudVPNGatewaysRead,

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

			"vpc_id": {
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

			"vpn_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"grade": {
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

						"charge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"auto_renew": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"vpc_id": {
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
					},
				},
			},
		},
	}
}

func dataSourceUCloudVPNGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ipsecvpnClient

	req := conn.NewDescribeVPNGatewayRequest()

	if v, ok := d.GetOk("ids"); ok {
		req.VPNGatewayIds = schemaSetToStringSlice(v)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	var allVPNGateways []ipsecvpn.VPNGatewayDataSet
	var vpnGateways []ipsecvpn.VPNGatewayDataSet
	var limit = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeVPNGateway(req)
		if err != nil {
			return fmt.Errorf("error on reading vpn gateway list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allVPNGateways = append(allVPNGateways, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allVPNGateways {
			if r != nil && !r.MatchString(v.VPNGatewayName) {
				continue
			}

			vpnGateways = append(vpnGateways, v)
		}
	} else {
		vpnGateways = allVPNGateways
	}

	err := dataSourceUCloudVPNGatewaysSave(d, vpnGateways)
	if err != nil {
		return fmt.Errorf("error on reading vpnGateway list, %s", err)
	}

	return nil
}

func dataSourceUCloudVPNGatewaysSave(d *schema.ResourceData, vpnGateways []ipsecvpn.VPNGatewayDataSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, vpnGateway := range vpnGateways {
		ids = append(ids, vpnGateway.VPNGatewayId)

		ipSet := []map[string]interface{}{}
		ipSet = append(ipSet, map[string]interface{}{
			"ip":            vpnGateway.EIP,
			"internet_type": vpnGateway.EIPType,
		})

		data = append(data, map[string]interface{}{
			"id":          vpnGateway.VPNGatewayId,
			"grade":       upperCamelCvt.convert(vpnGateway.Grade),
			"name":        vpnGateway.VPNGatewayName,
			"remark":      vpnGateway.Remark,
			"tag":         vpnGateway.Tag,
			"create_time": timestampToString(vpnGateway.CreateTime),
			"expire_time": timestampToString(vpnGateway.ExpireTime),
			"charge_type": upperCamelCvt.convert(vpnGateway.ChargeType),
			"auto_renew":  boolCamelCvt.unconvert(vpnGateway.AutoRenew),
			"ip_set":      ipSet,
			"vpc_id":      vpnGateway.VPCId,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("vpn_gateways", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
