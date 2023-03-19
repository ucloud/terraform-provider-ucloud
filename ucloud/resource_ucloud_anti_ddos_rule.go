package ucloud

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudAntiDDoSRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudAntiDDoSRuleCreate,
		Read:   resourceUCloudAntiDDoSRuleRead,
		Update: resourceUCloudAntiDDoSRuleUpdate,
		Delete: resourceUCloudAntiDDoSRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(),

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"port": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},

			"real_server_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"real_servers": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"toa_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  200,
			},

			"real_server_detection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  200,
			},

			"backup_server": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_index": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudAntiDDoSRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	req := conn.NewCreateBGPServiceFwdRuleRequest()
	req.ResourceId = ucloud.String(d.Get("instance_id").(string))
	req.Remark = ucloud.String(d.Get("comment").(string))
	ip := d.Get("ip").(string)
	port := 0
	req.BgpIP = ucloud.String(ip)
	if val, ok := d.GetOk("port"); ok {
		port = val.(int)
		req.BgpIPPort = ucloud.Int(port)
		req.FwdType = ucloud.String("TCP")
	} else {
		req.BgpIPPort = ucloud.Int(0)
		req.FwdType = ucloud.String("IP")
	}
	req.SourceType = ucloud.String(d.Get("real_server_type").(string))
	realServers := d.Get("real_servers").([]interface{})
	if len(realServers) > 1 {
		req.LoadBalance = ucloud.String("Yes")
	} else {
		req.LoadBalance = ucloud.String("No")
	}
	toaId := d.Get("toa_id").(int)
	for _, realServer := range realServers {
		realServerMap := realServer.(map[string]interface{})
		req.SourceAddrArr = append(req.SourceAddrArr, realServerMap["address"].(string))
		if port, ok := realServerMap["port"]; ok {
			req.SourcePortArr = append(req.SourcePortArr, strconv.Itoa(port.(int)))
		} else {
			req.SourcePortArr = append(req.SourcePortArr, "0")
		}
		req.SourceToaIDArr = append(req.SourceToaIDArr, strconv.Itoa(toaId))
	}
	if d.Get("real_server_detection").(bool) {
		req.SourceDetect = ucloud.Int(1)
		if backupServer, ok := d.GetOk("backup_server"); !ok {
			return errors.New("backup_server must be set when real_server_detection is true")
		} else {
			backupServerMap := backupServer.(map[string]interface{})
			req.BackupIP = ucloud.String(backupServerMap["ip"].(string))
			if port, portOk := backupServerMap["port"]; portOk {
				req.BackupPort = ucloud.Int(port.(int))
			} else {
				req.BackupPort = ucloud.Int(0)
			}
		}
	} else {
		req.SourceDetect = ucloud.Int(0)
	}
	resp, err := conn.CreateBGPServiceFwdRule(req)
	if err != nil {
		return fmt.Errorf("error on creating ucloud_anti_ddos_rule, %s", err)
	}
	instanceId := d.Get("instance_id").(string)
	idItems := []string{instanceId, ip}
	if port != 0 {
		idItems = append(idItems, strconv.Itoa(port))
	}
	d.SetId(strings.Join(idItems, "/"))
	d.Set("rule_index", resp.RuleIndex)
	// after create lb, we need to wait it initialized
	stateConf := antiDDoSRuleWaitForState(client, instanceId, resp.RuleIndex)

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for ucloud_anti_ddos_instance %q creating, %s", d.Id(), err)
	}

	return resourceUCloudAntiDDoSRuleRead(d, meta)
}

func resourceUCloudAntiDDoSRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	d.Partial(true)
	instanceId := d.Get("instance_id").(string)
	ruleIndex := d.Get("rule_index").(int)
	ruleId := d.Get("rule_id").(string)
	if d.HasChange("comment") && !d.IsNewResource() {
		d.SetPartial("comment")
		req := conn.NewSetNapFwdRuleRemarkRequest()
		req.ResourceId = ucloud.String(instanceId)
		req.RuleIndex = ucloud.String(strconv.Itoa(ruleIndex))
		req.Remark = ucloud.String(d.Get("comment").(string))
		_, err := conn.SetNapFwdRuleRemark(req)
		if err != nil {
			return fmt.Errorf("fail to set comment for rule %v of %v", ruleIndex, instanceId)
		}
	}
	if d.HasChanges("real_servers", "toa_id", "real_server_detection", "backup_server") && !d.IsNewResource() {
		d.SetPartial("real_servers")
		d.SetPartial("toa_id")
		d.SetPartial("real_server_detection")
		d.SetPartial("backup_server")

		req := conn.NewUpdateBGPServiceFwdRuleRequest()
		req.ResourceId = ucloud.String(instanceId)
		req.RuleID = ucloud.String(ruleId)
		req.RuleIndex = ucloud.Int(ruleIndex)
		req.BgpIP = ucloud.String(d.Get("ip").(string))
		if port, ok := d.GetOk("port"); ok {
			req.BgpIPPort = ucloud.Int(port.(int))
			req.FwdType = ucloud.String("TCP")
		} else {
			req.BgpIPPort = ucloud.Int(0)
			req.FwdType = ucloud.String("IP")
		}
		realServers := d.Get("real_servers").([]interface{})
		if len(realServers) > 1 {
			req.LoadBalance = ucloud.String("Yes")
		} else {
			req.LoadBalance = ucloud.String("No")
		}
		toaId := d.Get("toa_id").(int)
		for _, realServer := range realServers {
			realServerMap := realServer.(map[string]interface{})
			req.SourceAddrArr = append(req.SourceAddrArr, realServerMap["address"].(string))
			if port, ok := realServerMap["port"]; ok {
				req.SourcePortArr = append(req.SourcePortArr, strconv.Itoa(port.(int)))
			} else {
				req.SourcePortArr = append(req.SourcePortArr, "0")
			}
			req.SourceToaIDArr = append(req.SourceToaIDArr, strconv.Itoa(toaId))
		}
		if d.Get("real_server_detection").(bool) {
			req.SourceDetect = ucloud.Int(1)
			if backupServer, ok := d.GetOk("backup_server"); !ok {
				return errors.New("backup_server must be set when real_server_detection is true")
			} else {
				backupServerMap := backupServer.(map[string]interface{})
				req.BackupIP = ucloud.String(backupServerMap["ip"].(string))
				if port, portOk := backupServerMap["port"]; portOk {
					req.BackupPort = ucloud.Int(port.(int))
				} else {
					req.BackupPort = ucloud.Int(0)
				}
			}
		} else {
			req.SourceDetect = ucloud.Int(0)
		}
		_, err := conn.UpdateBGPServiceFwdRule(req)
		if err != nil {
			return fmt.Errorf("fail to update ucloud_anti_ddos_rule, %v", err)
		}
		stateConf := antiDDoSRuleWaitForState(client, instanceId, ruleIndex)

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error on waiting for ucloud_anti_ddos_instance %q updating, %s", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudAntiDDoSRuleRead(d, meta)
}

func resourceUCloudAntiDDoSRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	id := d.Id()
	idItems := strings.Split(id, "/")
	if len(idItems) > 3 || len(idItems) < 2 {
		return fmt.Errorf("%v is an invalid ucloud_anti_ddos_rule id", id)
	}

	instanceId := idItems[0]
	ip := idItems[1]
	port := 0
	if len(idItems) == 3 {
		var err error
		port, err = strconv.Atoi(idItems[2])
		if err != nil {
			return fmt.Errorf("%v is an invalid ucloud_anti_ddos_rule id, %v", id, err)
		}
	}

	ruleInfo, err := client.describeUADSBGPServiceFwdRuleByIpPort(instanceId, ip, port)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on ucloud_anti_ddos_rule of %s, %s", instanceId, err)
	}
	d.Set("ip", ruleInfo.BgpIP)
	d.Set("port", ruleInfo.BgpIPPort)
	d.Set("real_server_type", ruleInfo.SourceInfo.Type)
	var toaId int
	var realServers []interface{}
	for _, c := range ruleInfo.SourceInfo.Conf {
		server := map[string]interface{}{"address": c.Source}
		if c.Port != 0 {
			server["port"] = c.Port
		}
		realServers = append(realServers, server)
		toaId = c.Toa
	}
	d.Set("toa_id", toaId)
	if ruleInfo.SourceDetect != 0 {
		d.Set("real_server_detection", true)
		d.Set("backup_server", map[string]interface{}{"ip": ruleInfo.BackupIP, "port": ruleInfo.BackupPort})
	} else {
		d.Set("real_server_detection", false)
	}
	d.Set("comment", ruleInfo.Remark)
	d.Set("status", ruleInfo.Status)
	d.Set("rule_id", ruleInfo.RuleID)
	d.Set("rule_index", ruleInfo.RuleIndex)
	d.Set("real_servers", realServers)
	d.Set("instance_id", instanceId)
	return nil
}

func resourceUCloudAntiDDoSRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	instanceId := d.Get("instance_id").(string)
	ruleIndex := d.Get("rule_index").(int)

	req := conn.NewDeleteBGPServiceFwdRuleRequest()
	req.ResourceId = ucloud.String(instanceId)
	req.RuleIndex = ucloud.Int(ruleIndex)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteBGPServiceFwdRule(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ucloud_anti_ddos_rule %s, %s", d.Id(), err))
		}

		_, err := client.describeUADSBGPServiceFwdRule(instanceId, ruleIndex)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ucloud_anti_ddos_rule when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified ucloud_anti_ddos_rule %s has not been deleted due to unknown error", d.Id()))
	})
}

func antiDDoSRuleWaitForState(client *UCloudClient, id string, ruleIndex int) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			ruleInfo, err := client.describeUADSBGPServiceFwdRule(id, ruleIndex)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			if ruleInfo.Status == uadsBGPServiceFwdRuleStatusPending {
				return ruleInfo, statusPending, nil
			} else if ruleInfo.Status == uadsBGPServiceFwdRuleStatusSuccess {
				return ruleInfo, statusInitialized, nil
			} else {
				return nil, "", fmt.Errorf("status %v is unknown", ruleInfo.Status)
			}
		},
	}
}
