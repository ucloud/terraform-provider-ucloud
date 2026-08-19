package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

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
				Required:     true,
				ValidateFunc: validateName,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// read only view of the rules belong to this sec group, the rules
			// themselves are managed by the ucloud_sec_group_rule resource
			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"protocol_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"dst_port": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ip_range": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"rule_action": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceUCloudSecGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewCreateSecGroupRequest()
	req.Name = ucloud.String(d.Get("name").(string))
	req.VPCID = ucloud.String(d.Get("vpc_id").(string))

	resp, err := conn.CreateSecGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating sec group, %s", err)
	}

	d.SetId(resp.SecGroupId)

	// CreateSecGroup accepts neither tag nor remark, both can only be set
	// afterwards. the tag is always sent so that the remote value matches the
	// schema default instead of whatever the remote api picks on its own.
	tag := defaultTag
	if v, ok := d.GetOk("tag"); ok {
		tag = v.(string)
	}

	var remark string
	if v, ok := d.GetOk("remark"); ok {
		remark = v.(string)
	}

	if err := updateSecGroupAttribute(client, d.Id(), d.Get("name").(string), tag, remark); err != nil {
		return err
	}

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	d.Partial(true)

	if !d.IsNewResource() && (d.HasChange("name") || d.HasChange("tag") || d.HasChange("remark")) {
		// if tag is empty string, use default tag
		tag := defaultTag
		if v, ok := d.GetOk("tag"); ok {
			tag = v.(string)
		}

		if err := updateSecGroupAttribute(client, d.Id(), d.Get("name").(string), tag, d.Get("remark").(string)); err != nil {
			return err
		}

		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")
	}

	d.Partial(false)

	return resourceUCloudSecGroupRead(d, meta)
}

func resourceUCloudSecGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	sgSet, err := client.describeSecGroupById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sec group %q, %s", d.Id(), err)
	}

	d.Set("name", sgSet.Name)
	d.Set("vpc_id", sgSet.VPCId)
	d.Set("tag", sgSet.Tag)
	d.Set("remark", sgSet.Remark)
	d.Set("type", sgSet.Type)
	d.Set("create_time", timestampToString(sgSet.CreateTime))

	rules := []map[string]interface{}{}
	for _, item := range sgSet.Rule {
		rules = append(rules, map[string]interface{}{
			"direction":     item.Direction,
			"protocol_type": item.ProtocolType,
			"dst_port":      item.DstPort,
			"ip_range":      item.IPRange,
			"rule_action":   item.RuleAction,
			"priority":      item.Priority,
			"remark":        item.Remark,
			"rule_id":       item.RuleId,
		})
	}

	if err := d.Set("rules", rules); err != nil {
		return err
	}

	return nil
}

func resourceUCloudSecGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteSecGroupRequest()
	req.SecGroupId = []string{d.Id()}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		// the sec group can not be deleted while it is still bound to any
		// resource, and the unbinding takes a while to take effect, so the
		// failure is treated as retryable
		if _, err := conn.DeleteSecGroup(req); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting sec group %q, %s", d.Id(), err))
		}

		_, err := client.describeSecGroupById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sec group when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sec group %q has not been deleted due to unknown error", d.Id()))
	})
}

// UpdateSecGroupRequest.SecGroupId is defined as *string by the sdk while the
// api backend expects a string array, and the sdk request carries no Tag field
// at all even though its own comment lists Tag as one of the accepted values,
// so the request is built by hand.
// the api requires at least one of Name, Tag and Remark to be set, and it
// ignores an empty value instead of clearing the field.
func updateSecGroupAttribute(client *UCloudClient, secGroupId, name, tag, remark string) error {
	if name == "" && tag == "" && remark == "" {
		return nil
	}

	payload := map[string]interface{}{
		"Action":     "UpdateSecGroup",
		"SecGroupId": []string{secGroupId},
	}

	if name != "" {
		payload["Name"] = name
	}

	if tag != "" {
		payload["Tag"] = tag
	}

	if remark != "" {
		payload["Remark"] = remark
	}

	req := client.genericClient.NewGenericRequest()
	if err := req.SetPayload(payload); err != nil {
		return fmt.Errorf("error on setting payload for %s to sec group %q, %s", "UpdateSecGroup", secGroupId, err)
	}

	if _, err := client.genericClient.GenericInvoke(req); err != nil {
		return fmt.Errorf("error on %s to sec group %q, %s", "UpdateSecGroup", secGroupId, err)
	}

	return nil
}
