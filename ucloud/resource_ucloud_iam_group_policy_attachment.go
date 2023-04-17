package ucloud

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMGroupPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMGroupPolicyAttachmentCreate,
		Read:   resourceUCloudIAMGroupPolicyAttachmentRead,
		Delete: resourceUCloudIAMGroupPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_urn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIAMGroupPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewAttachPoliciesToGroupRequest()
	groupName := d.Get("group_name").(string)
	policyURN := d.Get("policy_urn").(string)
	req.PolicyURNs = []string{policyURN}
	req.GroupName = ucloud.String(groupName)
	var projectID string
	if projectID = d.Get("project_id").(string); projectID != "" {
		req.ProjectID = ucloud.String(projectID)
		req.Scope = ucloud.String("Specified")
	} else {
		req.Scope = ucloud.String("Unspecified")
	}

	_, err := conn.AttachPoliciesToGroup(req)
	if err != nil {
		return fmt.Errorf("error on attach policy to group, %s", err)
	}
	d.SetId(buildUCloudIAMGroupPolicyAttachmentID(groupName, policyURN, projectID))
	return resourceUCloudIAMGroupPolicyAttachmentRead(d, meta)
}

func resourceUCloudIAMGroupPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	groupName, policyURN, projectID, err := extractUCloudIAMGroupPolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	attachment, err := client.describeIAMGroupPolicyAttachment(groupName, policyURN, projectID)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading group policy attachment %q, %s", d.Id(), err)
	}
	d.Set("group_name", groupName)
	d.Set("policy_urn", attachment.PolicyURN)
	d.Set("project_id", projectID)
	d.Set("create_time", timestampToString(attachment.AttachedAt))
	return nil
}

func resourceUCloudIAMGroupPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn
	req := conn.NewDetachPoliciesFromGroupRequest()
	groupName, policyURN, projectID, err := extractUCloudIAMGroupPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("fail to delete policy attachment: %v", err)
	}
	req.GroupName = ucloud.String(groupName)
	req.PolicyURNs = []string{policyURN}
	if projectID != "" {
		req.Scope = ucloud.String("Specified")
		req.ProjectID = ucloud.String(projectID)
	} else {
		req.Scope = ucloud.String("Unspecified")
	}
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DetachPoliciesFromGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on detaching policies from group %q, %s", d.Id(), err))
		}

		_, err := client.describeIAMGroupPolicyAttachment(groupName, policyURN, projectID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading group policy attachment when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified group policy attachment %q has not been deleted due to unknown error", d.Id()))
	})
}

func buildUCloudIAMGroupPolicyAttachmentID(groupName string, policyURN string, projectID string) string {
	if projectID == "" {
		return fmt.Sprintf(accountScopeAttachmentPrefix+"%s/%s", groupName, policyURN)
	} else {
		return fmt.Sprintf(projectScopeAttachmentPrefix+"%s/%s/%s", projectID, groupName, policyURN)
	}
}

func extractUCloudIAMGroupPolicyAttachmentID(id string) (groupName string, policyURN string, projectID string, err error) {
	if strings.HasPrefix(id, accountScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, accountScopeAttachmentPrefix), "/", 2)
		return items[0], items[1], "", nil
	} else if strings.HasPrefix(id, projectScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, projectScopeAttachmentPrefix), "/", 3)
		return items[1], items[2], items[0], nil
	} else {
		return "", "", "", errors.New("fail to parse id")
	}
}
