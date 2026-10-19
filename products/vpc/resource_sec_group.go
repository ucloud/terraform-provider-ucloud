package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

// resourceUCloudSecGroup manages a security group. Its rules are not part of
// this resource: the API gives every rule a RuleId of its own, so each one is
// managed by ucloud_sec_group_rule instead.
func resourceUCloudSecGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSecGroupCreate,
		Read:   resourceUCloudSecGroupRead,
		Update: resourceUCloudSecGroupUpdate,
		Delete: resourceUCloudSecGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// tag can only be read: neither CreateSecGroup nor UpdateSecGroup
			// accepts a Tag parameter, so exposing it as writable would leave
			// a permanent diff for any configuration that set it.
			"tag": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudSecGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating sec group, %s", err)
	}
	conn := client.vpcconn

	req := conn.NewCreateSecGroupRequest()
	req.VPCID = ucloud.String(d.Get("vpc_id").(string))
	if value, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(value.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-sec-group-"))
	}

	resp, err := conn.CreateSecGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group, %s", err)
	}
	d.SetId(resp.SecGroupId)

	if _, err := secGroupWaitForState(client, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for sec group %q complete creating, %s", d.Id(), err)
	}

	// CreateSecGroup carries a name and a VPC and nothing else, so a remark in
	// the configuration has nowhere to go on the way in and has to be pushed by
	// the update API once the group exists. Leaving it out is not a silent
	// no-op: the read below would record the empty remark the group was created
	// with, and every plan after that would show the configured remark as a
	// change that applying could never clear.
	if value, ok := d.GetOk("remark"); ok {
		updateReq := conn.NewUpdateSecGroupRequest()
		updateReq.SecGroupId = []string{d.Id()}
		updateReq.Remark = ucloud.String(value.(string))
		if _, err := conn.UpdateSecGroup(updateReq); err != nil {
			return fmt.Errorf("error on %s to sec group %q when creating, %s", "UpdateSecGroup", d.Id(), err)
		}
	}

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading sec group, %s", err)
	}
	secGroup, err := client.describeSecGroupById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sec group %q, %s", d.Id(), err)
	}

	_ = d.Set("name", secGroup.Name)
	_ = d.Set("vpc_id", secGroup.VPCId)
	_ = d.Set("tag", secGroup.Tag)
	_ = d.Set("type", secGroup.Type)
	_ = d.Set("remark", secGroup.Remark)
	_ = d.Set("create_time", timestampToString(secGroup.CreateTime))

	return nil
}

func resourceUCloudSecGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating sec group, %s", err)
	}
	conn := client.vpcconn

	d.Partial(true)

	// UpdateSecGroup accepts name and remark in one request, but each is sent on
	// its own so that a call carries only the field that changed. A request
	// built from both would also send the unchanged one's current value, which
	// is the value last read back rather than one the user asked for, and would
	// quietly overwrite anything changed out of band since that read.
	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewUpdateSecGroupRequest()
		req.SecGroupId = []string{d.Id()}
		req.Name = ucloud.String(d.Get("name").(string))
		if _, err := conn.UpdateSecGroup(req); err != nil {
			return fmt.Errorf("error on %s to sec group %q, %s", "UpdateSecGroup", d.Id(), err)
		}
		d.SetPartial("name")
	}

	if d.HasChange("remark") && !d.IsNewResource() {
		req := conn.NewUpdateSecGroupRequest()
		req.SecGroupId = []string{d.Id()}
		req.Remark = ucloud.String(d.Get("remark").(string))
		if _, err := conn.UpdateSecGroup(req); err != nil {
			return fmt.Errorf("error on %s to sec group %q, %s", "UpdateSecGroup", d.Id(), err)
		}
		d.SetPartial("remark")
	}

	d.Partial(false)

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting sec group, %s", err)
	}
	conn := client.vpcconn

	req := conn.NewDeleteSecGroupRequest()
	req.SecGroupId = []string{d.Id()}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		// A group that is already gone makes the delete a success rather than
		// an error, so repeated destroys stay idempotent.
		if _, err := client.describeSecGroupById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sec group when deleting %q, %s", d.Id(), err))
		}

		if _, err := conn.DeleteSecGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting sec group %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sec group %q has not been deleted due to unknown error", d.Id()))
	})
}

// secGroupWaitForState waits for a security group to become readable. The API
// exposes no status field, so this is an existence wait: statusPending means
// "not readable yet" and statusInitialized means "readable".
func secGroupWaitForState(client *productClient, secGroupID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			secGroup, err := client.describeSecGroupById(secGroupID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return secGroup, statusInitialized, nil
		},
	}
}
