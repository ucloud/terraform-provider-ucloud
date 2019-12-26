package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
)

func resourceUCloudIsolationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIsolationGroupCreate,
		Read:   resourceUCloudIsolationGroupRead,
		Delete: resourceUCloudIsolationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateIsolationGroupName,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIsolationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn

	req := conn.NewCreateIsolationGroupRequest()
	if v, ok := d.GetOk("name"); ok {
		req.GroupName = ucloud.String(v.(string))
	} else {
		req.GroupName = ucloud.String(resource.PrefixedUniqueId("tf-isolation-group-"))
	}

	if v, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(v.(string))
	}

	resp, err := conn.CreateIsolationGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating isolation group, %s", err)
	}

	d.SetId(resp.GroupId)
	return resourceUCloudIsolationGroupRead(d, meta)
}

func resourceUCloudIsolationGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	igSet, err := client.describeIsolationGroupById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading isolation group %q, %s", d.Id(), err)
	}

	d.Set("name", igSet.GroupName)
	d.Set("remark", igSet.Remark)
	return nil
}

func resourceUCloudIsolationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn

	req := conn.NewDeleteIsolationGroupRequest()
	req.GroupId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteIsolationGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting isolation group %q, %s", d.Id(), err))
		}

		_, err := client.describeIsolationGroupById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading isolation group when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified isolation group %q has not been deleted due to unknown error", d.Id()))
	})
}
