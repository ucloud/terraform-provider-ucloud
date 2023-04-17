package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMGroupCreate,
		Update: resourceUCloudIAMGroupUpdate,
		Read:   resourceUCloudIAMGroupRead,
		Delete: resourceUCloudIAMGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceUCloudIAMGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewCreateGroupRequest()
	req.GroupName = ucloud.String(d.Get("name").(string))
	if val, ok := d.GetOk("comment"); ok {
		req.Description = ucloud.String(val.(string))
	}
	_, err := conn.CreateGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating group, %s", err)
	}
	d.SetId(d.Get("name").(string))
	return resourceUCloudIAMGroupRead(d, meta)
}

func resourceUCloudIAMGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewUpdateGroupRequest()
	req.GroupName = ucloud.String(d.Get("name").(string))

	if d.HasChange("comment") {
		req.Description = ucloud.String(d.Get("comment").(string))
		_, err := conn.UpdateGroup(req)
		if err != nil {
			return fmt.Errorf("error on update group, %s", err)
		}
	}
	return resourceUCloudIAMGroupRead(d, meta)
}

func resourceUCloudIAMGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	resp, err := client.describeGroup(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading group %q, %s", d.Id(), err)
	}
	d.Set("name", resp.GroupName)
	d.Set("comment", resp.Description)
	return nil
}

func resourceUCloudIAMGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewDeleteGroupRequest()
	req.GroupName = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting group %q, %s", d.Id(), err))
		}

		_, err := client.describeGroup(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading group when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified group %q has not been deleted due to unknown error", d.Id()))
	})
}
