package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"strconv"
	"time"
)

func resourceUCloudVPNConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPNConnectionCreate,
		Read:   resourceUCloudVPNConnectionRead,
		Update: resourceUCloudVPNConnectionUpdate,
		Delete: resourceUCloudVPNConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"customer_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateName,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ike_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ike_version": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "ikev1",
							ValidateFunc: validation.StringInSlice([]string{
								"ikev1",
							}, false),
						},

						"pre_shared_key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateVPNPreSharedKey,
						},

						"exchange_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "main",
							ValidateFunc: validation.StringInSlice([]string{
								"main",
								"aggressive",
							}, false),
						},

						"encryption_algorithm": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "aes128",
							ValidateFunc: validation.StringInSlice([]string{
								"aes128",
								"aes192",
								"aes256",
								"aes512",
								"3des",
							}, false),
						},

						"authentication_algorithm": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "sha1",
							ValidateFunc: validation.StringInSlice([]string{
								"md5",
								"sha1",
								"sha2-256",
							}, false),
						},

						"local_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateVpnAuto,
						},

						"remote_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateVpnAuto,
						},

						"dh_group": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "15",
							ValidateFunc: validation.StringInSlice([]string{
								"1",
								"2",
								"5",
								"14",
								"15",
								"16",
							}, false),
						},

						"sa_life_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      86400,
							ValidateFunc: validation.IntBetween(600, 604800),
						},
					},
				},
			},

			"ipsec_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"local_subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set:      schema.HashString,
							MinItems: 1,
							MaxItems: 10,
						},

						"remote_subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRBlock,
							},
							Set:      schema.HashString,
							MinItems: 1,
							MaxItems: 20,
						},

						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "esp",
							ValidateFunc: validation.StringInSlice([]string{
								"esp",
								"ah",
							}, false),
						},

						"encryption_algorithm": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "aes128",
							ValidateFunc: validation.StringInSlice([]string{
								"aes128",
								"aes192",
								"aes256",
								"aes512",
								"3des",
							}, false),
						},

						"authentication_algorithm": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "sha1",
							ValidateFunc: validation.StringInSlice([]string{
								"md5",
								"sha1",
							}, false),
						},

						"sa_life_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3600,
							ValidateFunc: validation.IntBetween(1200, 604800),
						},

						"sa_life_time_bytes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(8000, 20000000),
							Computed:     true,
						},

						"pfs_dh_group": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "disable",
							ValidateFunc: validation.StringInSlice([]string{
								"disable",
								"1",
								"2",
								"5",
								"14",
								"15",
								"16",
							}, false),
						},
					},
				},
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudVPNConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewCreateVPNTunnelRequest()
	req.VPNGatewayId = ucloud.String(d.Get("vpn_gateway_id").(string))
	req.RemoteVPNGatewayId = ucloud.String(d.Get("customer_gateway_id").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))

	if v, ok := d.GetOk("name"); ok {
		req.VPNTunnelName = ucloud.String(v.(string))
	} else {
		req.VPNTunnelName = ucloud.String(resource.PrefixedUniqueId("tf-vpn-connection-"))
	}
	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	if v, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(v.(string))
	}

	ikeCfg := d.Get("ike_config").([]interface{})[0].(map[string]interface{})
	req.IKEVersion = ucloud.String(vpnIkeVersionCvt.unconvert(ikeCfg["ike_version"].(string)))
	req.IKEPreSharedKey = ucloud.String(ikeCfg["pre_shared_key"].(string))
	req.IKEExchangeMode = ucloud.String(ikeCfg["exchange_mode"].(string))
	req.IKEEncryptionAlgorithm = ucloud.String(ikeCfg["encryption_algorithm"].(string))
	req.IKEAuthenticationAlgorithm = ucloud.String(ikeCfg["authentication_algorithm"].(string))
	req.IKEDhGroup = ucloud.String(ikeCfg["dh_group"].(string))
	req.IKESALifetime = ucloud.String(strconv.Itoa(ikeCfg["sa_life_time"].(int)))
	if ikeCfg["local_id"].(string) != "" {
		req.IKELocalId = ucloud.String(ikeCfg["local_id"].(string))
	} else {
		req.IKELocalId = ucloud.String("auto")
	}

	if ikeCfg["remote_id"].(string) != "" {
		req.IKERemoteId = ucloud.String(ikeCfg["remote_id"].(string))
	} else {
		req.IKERemoteId = ucloud.String("auto")
	}

	ipsecCfg := d.Get("ipsec_config").([]interface{})[0].(map[string]interface{})
	req.IPSecLocalSubnetIds = schemaSetToStringSlice(ipsecCfg["local_subnet_ids"].(*schema.Set))
	req.IPSecRemoteSubnets = schemaSetToStringSlice(ipsecCfg["remote_subnets"].(*schema.Set))

	req.IPSecProtocol = ucloud.String(ipsecCfg["protocol"].(string))
	req.IPSecEncryptionAlgorithm = ucloud.String(ipsecCfg["encryption_algorithm"].(string))
	req.IPSecAuthenticationAlgorithm = ucloud.String(ipsecCfg["authentication_algorithm"].(string))
	req.IPSecSALifetime = ucloud.String(strconv.Itoa(ipsecCfg["sa_life_time"].(int)))

	if ipsecCfg["sa_life_time_bytes"].(int) != 0 {
		req.IPSecSALifetimeBytes = ucloud.String(strconv.Itoa(ipsecCfg["sa_life_time_bytes"].(int)))
	}

	req.IPSecPFSDhGroup = ucloud.String(ipsecCfg["pfs_dh_group"].(string))

	resp, err := conn.CreateVPNTunnel(req)
	if err != nil {
		return fmt.Errorf("error on creating vpn connection, %s", err)
	}

	d.SetId(resp.VPNTunnelId)
	return resourceUCloudVPNConnectionRead(d, meta)
}

func resourceUCloudVPNConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient
	d.Partial(true)

	req := conn.NewUpdateVPNTunnelAttributeRequest()
	req.VPNTunnelId = ucloud.String(d.Id())

	updateAttribute := false
	if d.HasChange("ike_config") && !d.IsNewResource() {
		cfg := d.Get("ike_config").([]interface{})[0].(map[string]interface{})
		req.IKEPreSharedKey = ucloud.String(cfg["pre_shared_key"].(string))
		req.IKEExchangeMode = ucloud.String(cfg["exchange_mode"].(string))
		req.IKEEncryptionAlgorithm = ucloud.String(cfg["encryption_algorithm"].(string))
		req.IKEAuthenticationAlgorithm = ucloud.String(cfg["authentication_algorithm"].(string))
		req.IKEDhGroup = ucloud.String(cfg["dh_group"].(string))
		req.IKESALifetime = ucloud.String(strconv.Itoa(cfg["sa_life_time"].(int)))
		if cfg["local_id"].(string) != "" {
			req.IKELocalId = ucloud.String(cfg["local_id"].(string))
		}

		if cfg["remote_id"].(string) != "" {
			req.IKERemoteId = ucloud.String(cfg["remote_id"].(string))
		}
		updateAttribute = true
	}

	if d.HasChange("ipsec_config") && !d.IsNewResource() {
		cfg := d.Get("ipsec_config").([]interface{})[0].(map[string]interface{})
		req.IPSecLocalSubnetIds = schemaSetToStringSlice(cfg["local_subnet_ids"].(*schema.Set))
		req.IPSecRemoteSubnets = schemaSetToStringSlice(cfg["remote_subnets"].(*schema.Set))

		req.IPSecProtocol = ucloud.String(cfg["protocol"].(string))
		req.IPSecEncryptionAlgorithm = ucloud.String(cfg["encryption_algorithm"].(string))
		req.IPSecAuthenticationAlgorithm = ucloud.String(cfg["authentication_algorithm"].(string))
		req.IPSecSALifetime = ucloud.String(strconv.Itoa(cfg["sa_life_time"].(int)))
		req.IPSecPFSDhGroup = ucloud.String(cfg["pfs_dh_group"].(string))

		if cfg["sa_life_time_bytes"].(int) != 0 {
			req.IPSecSALifetimeBytes = ucloud.String(strconv.Itoa(cfg["sa_life_time_bytes"].(int)))
		}

		updateAttribute = true
	}

	if updateAttribute {
		if _, err := conn.UpdateVPNTunnelAttribute(req); err != nil {
			return fmt.Errorf("error on %s to vpn connection %q, %s", "UpdateVPNTunnelAttribute", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudVPNConnectionRead(d, meta)
}
func resourceUCloudVPNConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	vcSet, err := client.describeVPNConnectionById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpn connection %q, %s", d.Id(), err)
	}

	d.Set("name", vcSet.VPNTunnelName)
	d.Set("remark", vcSet.Remark)
	d.Set("tag", vcSet.Tag)
	d.Set("vpc_id", vcSet.VPCId)
	d.Set("vpn_gateway_id", vcSet.VPNGatewayId)
	d.Set("customer_gateway_id", vcSet.RemoteVPNGatewayId)
	d.Set("create_time", timestampToString(vcSet.CreateTime))

	ikeData := map[string]interface{}{
		"ike_version":              vpnIkeVersionCvt.convert(vcSet.IKEData.IKEVersion),
		"pre_shared_key":           vcSet.IKEData.IKEPreSharedKey,
		"exchange_mode":            vcSet.IKEData.IKEExchangeMode,
		"encryption_algorithm":     vcSet.IKEData.IKEEncryptionAlgorithm,
		"authentication_algorithm": vcSet.IKEData.IKEAuthenticationAlgorithm,
		"local_id":                 vpnAutoCvt.convert(vcSet.IKEData.IKELocalId),
		"remote_id":                vpnAutoCvt.convert(vcSet.IKEData.IKERemoteId),
		"dh_group":                 vcSet.IKEData.IKEDhGroup,
	}

	if v, err := strconv.Atoi(vcSet.IKEData.IKESALifetime); err != nil {
		return err
	} else {
		ikeData["sa_life_time"] = v
	}

	ikeConfig := []map[string]interface{}{}
	ikeConfig = append(ikeConfig, ikeData)
	if err := d.Set("ike_config", ikeConfig); err != nil {
		return err
	}

	ipsecData := map[string]interface{}{
		"local_subnet_ids":         vcSet.IPSecData.IPSecLocalSubnetIds,
		"remote_subnets":           vcSet.IPSecData.IPSecRemoteSubnets,
		"protocol":                 vcSet.IPSecData.IPSecProtocol,
		"encryption_algorithm":     vcSet.IPSecData.IPSecEncryptionAlgorithm,
		"authentication_algorithm": vcSet.IPSecData.IPSecAuthenticationAlgorithm,
		"pfs_dh_group":             vpnDisableCvt.convert(vcSet.IPSecData.IPSecPFSDhGroup),
	}

	if v, err := strconv.Atoi(vcSet.IPSecData.IPSecSALifetime); err != nil {
		return err
	} else {
		ipsecData["sa_life_time"] = v
	}

	if vcSet.IPSecData.IPSecSALifetimeBytes != "" {
		if v, err := strconv.Atoi(vcSet.IPSecData.IPSecSALifetimeBytes); err != nil {
			return err
		} else {
			ipsecData["sa_life_time_bytes"] = v
		}
	}

	ipsecConfig := []map[string]interface{}{}
	ipsecConfig = append(ipsecConfig, ipsecData)
	if err := d.Set("ipsec_config", ipsecConfig); err != nil {
		return err
	}

	return nil
}

func resourceUCloudVPNConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ipsecvpnClient

	req := conn.NewDeleteVPNTunnelRequest()
	req.VPNTunnelId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPNTunnel(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpn connection %q, %s", d.Id(), err))
		}

		_, err := client.describeVPNConnectionById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpn connection when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vpn connection %q has not been deleted due to unknown error", d.Id()))
	})
}
