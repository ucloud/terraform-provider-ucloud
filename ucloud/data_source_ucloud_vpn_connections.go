package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dateSourceUCloudVPNConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudVPNConnectionsRead,

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

			"vpn_connections": {
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

						"vpn_gateway_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"customer_gateway_id": {
							Type:     schema.TypeString,
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

						"ike_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ike_version": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"pre_shared_key": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"exchange_mode": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"encryption_algorithm": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"authentication_algorithm": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"local_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"remote_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"dh_group": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"sa_life_time": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"ipsec_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"local_subnet_ids": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},

									"remote_subnets": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},

									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"encryption_algorithm": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"authentication_algorithm": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"sa_life_time": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"sa_life_time_bytes": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"pfs_dh_group": {
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

func dataSourceUCloudVPNConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ipsecvpnClient

	req := conn.NewDescribeVPNTunnelRequest()

	if v, ok := d.GetOk("ids"); ok {
		req.VPNTunnelIds = schemaSetToStringSlice(v)
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	var allVPNConnections []ipsecvpn.VPNTunnelDataSet
	var vpnConnections []ipsecvpn.VPNTunnelDataSet
	var limit = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeVPNTunnel(req)
		if err != nil {
			return fmt.Errorf("error on reading vpn connection list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allVPNConnections = append(allVPNConnections, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allVPNConnections {
			if r != nil && !r.MatchString(v.VPNTunnelName) {
				continue
			}

			vpnConnections = append(vpnConnections, v)
		}
	} else {
		vpnConnections = allVPNConnections
	}

	err := dataSourceUCloudVPNConnectionsSave(d, vpnConnections)
	if err != nil {
		return fmt.Errorf("error on reading vpn connection list, %s", err)
	}

	return nil
}

func dataSourceUCloudVPNConnectionsSave(d *schema.ResourceData, vpnConnections []ipsecvpn.VPNTunnelDataSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, vpnConnection := range vpnConnections {
		ids = append(ids, vpnConnection.VPNTunnelId)
		ikeData := []map[string]interface{}{}
		ikeData = append(ikeData, map[string]interface{}{
			"ike_version":              vpnIkeVersionCvt.convert(vpnConnection.IKEData.IKEVersion),
			"pre_shared_key":           vpnConnection.IKEData.IKEPreSharedKey,
			"exchange_mode":            vpnConnection.IKEData.IKEExchangeMode,
			"encryption_algorithm":     vpnConnection.IKEData.IKEEncryptionAlgorithm,
			"authentication_algorithm": vpnConnection.IKEData.IKEAuthenticationAlgorithm,
			"local_id":                 vpnAutoCvt.convert(vpnConnection.IKEData.IKELocalId),
			"remote_id":                vpnAutoCvt.convert(vpnConnection.IKEData.IKERemoteId),
			"dh_group":                 vpnConnection.IKEData.IKEDhGroup,
		})
		if v, err := strconv.Atoi(vpnConnection.IKEData.IKESALifetime); err != nil {
			return err
		} else {
			ikeData[0]["sa_life_time"] = v
		}

		ipsecData := []map[string]interface{}{}
		ipsecData = append(ipsecData, map[string]interface{}{
			"local_subnet_ids":         vpnConnection.IPSecData.IPSecLocalSubnetIds,
			"remote_subnets":           vpnConnection.IPSecData.IPSecRemoteSubnets,
			"protocol":                 vpnConnection.IPSecData.IPSecProtocol,
			"encryption_algorithm":     vpnConnection.IPSecData.IPSecEncryptionAlgorithm,
			"authentication_algorithm": vpnConnection.IPSecData.IPSecAuthenticationAlgorithm,
			"pfs_dh_group":             vpnDisableCvt.convert(vpnConnection.IPSecData.IPSecPFSDhGroup),
		})

		if v, err := strconv.Atoi(vpnConnection.IPSecData.IPSecSALifetime); err != nil {
			return err
		} else {
			ipsecData[0]["sa_life_time"] = v
		}

		if vpnConnection.IPSecData.IPSecSALifetimeBytes != "" {
			if v, err := strconv.Atoi(vpnConnection.IPSecData.IPSecSALifetimeBytes); err != nil {
				return err
			} else {
				ipsecData[0]["sa_life_time_bytes"] = v
			}
		}

		data = append(data, map[string]interface{}{
			"id":                  vpnConnection.VPNTunnelId,
			"name":                vpnConnection.VPNTunnelName,
			"remark":              vpnConnection.Remark,
			"tag":                 vpnConnection.Tag,
			"vpc_id":              vpnConnection.VPCId,
			"create_time":         timestampToString(vpnConnection.CreateTime),
			"vpn_gateway_id":      vpnConnection.VPNGatewayId,
			"customer_gateway_id": vpnConnection.RemoteVPNGatewayId,
			"ike_config":          ikeData,
			"ipsec_config":        ipsecData,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("vpn_connections", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
