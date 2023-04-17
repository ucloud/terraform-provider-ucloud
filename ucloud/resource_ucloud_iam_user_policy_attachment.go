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

func resourceUCloudIAMUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMUserPolicyAttachmentCreate,
		Read:   resourceUCloudIAMUserPolicyAttachmentRead,
		Delete: resourceUCloudIAMUserPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"user_name": {
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

func resourceUCloudIAMUserPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewAttachPoliciesToUserRequest()
	userName := d.Get("user_name").(string)
	policyURN := d.Get("policy_urn").(string)
	req.PolicyURNs = []string{policyURN}
	req.UserName = ucloud.String(userName)
	var projectID string
	if projectID = d.Get("project_id").(string); projectID != "" {
		req.ProjectID = ucloud.String(projectID)
		req.Scope = ucloud.String("Specified")
	} else {
		req.Scope = ucloud.String("Unspecified")
	}

	_, err := conn.AttachPoliciesToUser(req)
	if err != nil {
		return fmt.Errorf("error on attach policy to user, %s", err)
	}
	d.SetId(buildUCloudIAMUserPolicyAttachmentID(userName, policyURN, projectID))
	return resourceUCloudIAMUserPolicyAttachmentRead(d, meta)
}

func resourceUCloudIAMUserPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	userName, policyURN, projectID, err := extractUCloudIAMUserPolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	attachment, err := client.describeIAMUserPolicyAttachment(userName, policyURN, projectID)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading user policy attachment %q, %s", d.Id(), err)
	}
	d.Set("user_name", userName)
	d.Set("policy_urn", attachment.PolicyURN)
	d.Set("project_id", attachment.ProjectID)
	d.Set("create_time", timestampToString(attachment.AttachedAt))
	return nil
}

func resourceUCloudIAMUserPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn
	req := conn.NewDetachPoliciesFromUserRequest()
	userName, policyURN, projectID, err := extractUCloudIAMUserPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("fail to delete policy attachment: %v", err)
	}
	req.UserName = ucloud.String(userName)
	req.PolicyURNs = []string{policyURN}
	if projectID != "" {
		req.Scope = ucloud.String("Specified")
		req.ProjectID = ucloud.String(projectID)
	} else {
		req.Scope = ucloud.String("Unspecified")
	}
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DetachPoliciesFromUser(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on detaching policies from user %q, %s", d.Id(), err))
		}

		_, err := client.describeIAMUserPolicyAttachment(userName, policyURN, projectID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading user policy attachment when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified user policy attachment %q has not been deleted due to unknown error", d.Id()))
	})
}

const accountScopeAttachmentPrefix = "account/"
const projectScopeAttachmentPrefix = "project/"

func buildUCloudIAMUserPolicyAttachmentID(userName string, policyURN string, projectID string) string {
	if projectID == "" {
		return fmt.Sprintf(accountScopeAttachmentPrefix+"%s/%s", userName, policyURN)
	} else {
		return fmt.Sprintf(projectScopeAttachmentPrefix+"%s/%s/%s", projectID, userName, policyURN)
	}
}

func extractUCloudIAMUserPolicyAttachmentID(id string) (userName string, policyURN string, projectID string, err error) {
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
