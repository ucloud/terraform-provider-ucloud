package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMProjectCreate,
		Update: resourceUCloudIAMProjectUpdate,
		Read:   resourceUCloudIAMProjectRead,
		Delete: resourceUCloudIAMProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIAMProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewCreateProjectRequest()
	req.ProjectName = ucloud.String(d.Get("name").(string))
	resp, err := conn.CreateProject(req)
	if err != nil {
		return fmt.Errorf("error on creating user, %s", err)
	}
	d.SetId(resp.ProjectId)
	return resourceUCloudIAMProjectRead(d, meta)
}

func resourceUCloudIAMProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn
	d.Partial(true)

	req := conn.NewModifyProjectRequest()
	req.ProjectId = ucloud.String(d.Id())
	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.ProjectName = ucloud.String(name)
		_, err := conn.ModifyProject(req)
		if err != nil {
			return fmt.Errorf("error on %s to update project, %s", d.Id(), err)
		}
		d.SetPartial("name")
	}
	d.Partial(false)
	return resourceUCloudIAMProjectRead(d, meta)
}

func resourceUCloudIAMProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	project, err := client.describeIAMProjectById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading project %q, %s", d.Id(), err)
	}
	d.Set("name", project.ProjectName)
	d.Set("create_time", timestampToString(project.CreatedAt))
	return nil
}

func resourceUCloudIAMProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewDeleteProjectRequest()
	req.ProjectID = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteProject(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting project %q, %s", d.Id(), err))
		}

		_, err := client.describeIAMProjectById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading project when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified project %q has not been deleted due to unknown error", d.Id()))
	})
}
