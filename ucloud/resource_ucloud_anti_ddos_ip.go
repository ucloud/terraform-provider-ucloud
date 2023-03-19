package ucloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudAntiDDoSIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudAntiDDoSIPCreate,
		Read:   resourceUCloudAntiDDoSIPRead,
		Update: resourceUCloudAntiDDoSIPUpdate,
		Delete: resourceUCloudAntiDDoSIPDelete,
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

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudAntiDDoSIPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	req := conn.NewCreateBGPServiceIPRequest()
	req.ResourceId = ucloud.String(d.Get("instance_id").(string))
	req.Remark = ucloud.String(d.Get("comment").(string))

	resp, err := conn.CreateBGPServiceIP(req)
	if err != nil {
		return fmt.Errorf("error on creating ucloud_anti_ddos_ip, %s", err)
	}
	instanceId := d.Get("instance_id").(string)
	d.SetId(fmt.Sprintf("%s/%s", instanceId, resp.DefenceIP))
	d.Set("ip", resp.DefenceIP)
	// after create lb, we need to wait it initialized
	stateConf := antiDDoSIPWaitForState(client, instanceId, resp.DefenceIP)

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for ucloud_anti_ddos_instance %q creating, %s", d.Id(), err)
	}

	return resourceUCloudAntiDDoSIPRead(d, meta)
}

func resourceUCloudAntiDDoSIPUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	d.Partial(true)
	instanceId := d.Get("instance_id").(string)
	ip := d.Get("ip").(string)

	if d.HasChange("comment") && !d.IsNewResource() {
		comment := d.Get("comment").(string)
		remarkReq := conn.NewSetNapIpRemarkRequest()
		remarkReq.ResourceId = ucloud.String(instanceId)
		remarkReq.NapIp = ucloud.String(ip)
		remarkReq.Remark = ucloud.String(comment)
		_, err := conn.SetNapIpRemark(remarkReq)
		if err != nil {
			return fmt.Errorf("fail to set comment %v for ip %v of %v", comment, ip, instanceId)
		}
		d.SetPartial("comment")
	}

	d.Partial(false)

	return resourceUCloudAntiDDoSIPRead(d, meta)
}

func resourceUCloudAntiDDoSIPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	id := d.Id()
	items := strings.SplitN(id, "/", 2)
	if len(items) != 2 {
		return fmt.Errorf("%v is an invalid ucloud_anti_ddos_ip id", id)
	}
	instanceId := items[0]
	ip := items[1]

	ipInfo, err := client.describeUADSBGPServiceIP(instanceId, ip)

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading uads ip %s of %s, %s", ip, instanceId, err)
	}
	d.Set("ip", ip)
	d.Set("instance_id", instanceId)
	d.Set("comment", ipInfo.Remark)
	d.Set("status", ipInfo.Status)
	d.Set("domain", ipInfo.Cname)
	return nil
}

func resourceUCloudAntiDDoSIPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	instanceId := d.Get("instance_id").(string)
	ip := d.Get("ip").(string)

	req := conn.NewDeleteBGPServiceIPRequest()
	req.ResourceId = ucloud.String(instanceId)
	req.DefenceIp = ucloud.String(ip)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteBGPServiceIP(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ucloud_anti_ddos_ip %s, %s", d.Id(), err))
		}

		_, err := client.describeUADSBGPServiceIP(instanceId, ip)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ucloud_anti_ddos_ip when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified ucloud_anti_ddos_ip %s has not been deleted due to unknown error", d.Id()))
	})
}

func antiDDoSIPWaitForState(client *UCloudClient, id string, ip string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			ipInfo, err := client.describeUADSBGPServiceIP(id, ip)
			if err != nil {
				log.Fatal(err)
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return ipInfo, "", err
			}
			if ipInfo.Status == uadsBGPServiceIPStatusPending {
				return ipInfo, statusPending, nil
			} else if ipInfo.Status == uadsBGPServiceIPStatusSuccess {
				return ipInfo, statusInitialized, nil
			} else {
				return nil, "", fmt.Errorf("status %v is unknown", ipInfo.Status)
			}
		},
	}
}
