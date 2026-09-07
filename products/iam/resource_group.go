package iam

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
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewCreateGroupRequest()
	req.GroupName = ucloud.String(d.Get("name").(string))
	if value, ok := d.GetOk("comment"); ok {
		req.Description = ucloud.String(value.(string))
	}
	if _, err := client.iamconn.CreateGroup(req); err != nil {
		return fmt.Errorf("error on creating group, %s", err)
	}
	d.SetId(d.Get("name").(string))
	return resourceUCloudIAMGroupRead(d, meta)
}

func resourceUCloudIAMGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewUpdateGroupRequest()
	req.GroupName = ucloud.String(d.Get("name").(string))

	if d.HasChange("comment") {
		req.Description = ucloud.String(d.Get("comment").(string))
		if _, err := client.iamconn.UpdateGroup(req); err != nil {
			return fmt.Errorf("error on update group, %s", err)
		}
	}
	return resourceUCloudIAMGroupRead(d, meta)
}

func resourceUCloudIAMGroupRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	group, err := client.describeGroup(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading group %q, %s", d.Id(), err)
	}
	_ = d.Set("name", group.GroupName)
	_ = d.Set("comment", group.Description)
	return nil
}

func resourceUCloudIAMGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewDeleteGroupRequest()
	req.GroupName = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.iamconn.DeleteGroup(req); err != nil {
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
