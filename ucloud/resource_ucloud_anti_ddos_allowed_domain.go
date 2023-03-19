package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudAntiDDoSAllowedDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudAntiDDoSAllowedDomainCreate,
		Read:   resourceUCloudAntiDDoSAllowedDomainRead,
		Update: resourceUCloudAntiDDoSAllowedDomainUpdate,
		Delete: resourceUCloudAntiDDoSAllowedDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(),

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudAntiDDoSAllowedDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	domain := d.Get("domain").(string)
	req := conn.NewAddNapAllowListDomainRequest()
	req.ResourceId = ucloud.String(d.Get("instance_id").(string))
	req.Domain = []string{domain}

	resp, err := conn.AddNapAllowListDomain(req)
	if err != nil {
		return fmt.Errorf("error on creating ucloud_anti_ddos_domain, %s", err)
	}

	for _, data := range resp.Data {
		if data.Domain == domain && data.Code != 0 {
			return fmt.Errorf("fail to create domain %v to %v with RetCode %v", data.Domain, d.Id(), data.Code)
		}
	}
	instanceId := d.Get("instance_id").(string)
	d.SetId(fmt.Sprintf("%s/%s", instanceId, d.Get("domain").(string)))

	// after create lb, we need to wait it initialized
	stateConf := antiDDoSAllowedDomainWaitForState(client, instanceId, domain)

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for ucloud_anti_ddos_instance %q creating, %s", d.Id(), err)
	}

	if comment, ok := d.GetOkExists("comment"); ok {
		remarkReq := conn.NewSetNapDomainEntryRemarkRequest()
		remarkReq.ResourceId = req.ResourceId
		remarkReq.Domain = ucloud.String(domain)
		remarkReq.Remark = ucloud.String(comment.(string))
		_, err := conn.SetNapDomainEntryRemark(remarkReq)
		if err != nil {
			return fmt.Errorf("fail to set comment %v for domain %v of %v", comment, domain, *req.ResourceId)
		}
	}

	return resourceUCloudAntiDDoSAllowedDomainRead(d, meta)
}

func resourceUCloudAntiDDoSAllowedDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	d.Partial(true)
	instanceId := d.Get("instance_id").(string)
	domain := d.Get("domain").(string)

	if d.HasChange("comment") && !d.IsNewResource() {
		comment := d.Get("comment").(string)
		remarkReq := conn.NewSetNapDomainEntryRemarkRequest()
		remarkReq.ResourceId = ucloud.String(instanceId)
		remarkReq.Domain = ucloud.String(domain)
		remarkReq.Remark = ucloud.String(comment)
		_, err := conn.SetNapDomainEntryRemark(remarkReq)
		if err != nil {
			return fmt.Errorf("fail to set comment %v for domain %v of %v", comment, domain, instanceId)
		}
		d.SetPartial("comment")
	}

	d.Partial(false)

	return resourceUCloudAntiDDoSAllowedDomainRead(d, meta)
}

func resourceUCloudAntiDDoSAllowedDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	id := d.Id()
	items := strings.SplitN(id, "/", 2)
	if len(items) != 2 {
		return fmt.Errorf("%v is an invalid ucloud_anti_ddos_allowed_domain id", id)
	}
	instanceId := items[0]
	domain := items[1]

	domainInfo, err := client.describeUADSAllowedDomain(instanceId, domain)

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading uads allowed domain %s of %s, %s", domain, instanceId, err)
	}
	d.Set("domain", domain)
	d.Set("instance_id", instanceId)
	d.Set("comment", domainInfo.Remark)
	d.Set("status", uadsAllowedDomainStatusCvt.convert(domainInfo.Status))
	return nil
}

func resourceUCloudAntiDDoSAllowedDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	instanceId := d.Get("instance_id").(string)
	domain := d.Get("domain").(string)

	req := conn.NewDeleteNapAllowListDomainRequest()
	req.ResourceId = ucloud.String(instanceId)
	req.Domain = []string{domain}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNapAllowListDomain(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ucloud_anti_ddos_allowed_domain %s, %s", d.Id(), err))
		}

		_, err := client.describeUADSAllowedDomain(instanceId, domain)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ucloud_anti_ddos_allowed_domain when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified ucloud_anti_ddos_allowed_domain %s has not been deleted due to unknown error", d.Id()))
	})
}

func antiDDoSAllowedDomainWaitForState(client *UCloudClient, id string, domain string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			domainInfo, err := client.describeUADSAllowedDomain(id, domain)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			if domainInfo.Status == uadsAllowedDomainStatusSuccess {
				return domainInfo, statusInitialized, nil
			} else if domainInfo.Status == uadsAllowedDomainStatusAdding {
				return domainInfo, statusPending, nil
			} else if domainInfo.Status == uadsAllowedDomainStatusFailure {
				return nil, "", fmt.Errorf("fail to add domain %v for %v", domain, id)
			} else {
				return nil, "", fmt.Errorf("status %v should not be got", domainInfo.Status)
			}
		},
	}
}
