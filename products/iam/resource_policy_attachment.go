package iam

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

const (
	accountScopeAttachmentPrefix = "account/"
	projectScopeAttachmentPrefix = "project/"
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewAttachPoliciesToUserRequest()
	userName := d.Get("user_name").(string)
	policyURN := d.Get("policy_urn").(string)
	req.PolicyURNs = []string{policyURN}
	req.UserName = ucloud.String(userName)
	projectID := d.Get("project_id").(string)
	if projectID != "" {
		req.ProjectID = ucloud.String(projectID)
		req.Scope = ucloud.String("Specified")
	} else {
		req.Scope = ucloud.String("Unspecified")
	}

	if _, err := client.iamconn.AttachPoliciesToUser(req); err != nil {
		return fmt.Errorf("error on attach policy to user, %s", err)
	}
	d.SetId(buildUCloudIAMUserPolicyAttachmentID(userName, policyURN, projectID))
	return resourceUCloudIAMUserPolicyAttachmentRead(d, meta)
}

func resourceUCloudIAMUserPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

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
	_ = d.Set("user_name", userName)
	_ = d.Set("policy_urn", attachment.PolicyURN)
	_ = d.Set("project_id", attachment.ProjectID)
	_ = d.Set("create_time", timestampToString(attachment.AttachedAt))
	return nil
}

func resourceUCloudIAMUserPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewDetachPoliciesFromUserRequest()
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
		if _, err := client.iamconn.DetachPoliciesFromUser(req); err != nil {
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewAttachPoliciesToGroupRequest()
	groupName := d.Get("group_name").(string)
	policyURN := d.Get("policy_urn").(string)
	req.PolicyURNs = []string{policyURN}
	req.GroupName = ucloud.String(groupName)
	projectID := d.Get("project_id").(string)
	if projectID != "" {
		req.ProjectID = ucloud.String(projectID)
		req.Scope = ucloud.String("Specified")
	} else {
		req.Scope = ucloud.String("Unspecified")
	}

	if _, err := client.iamconn.AttachPoliciesToGroup(req); err != nil {
		return fmt.Errorf("error on attach policy to group, %s", err)
	}
	d.SetId(buildUCloudIAMGroupPolicyAttachmentID(groupName, policyURN, projectID))
	return resourceUCloudIAMGroupPolicyAttachmentRead(d, meta)
}

func resourceUCloudIAMGroupPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

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
	_ = d.Set("group_name", groupName)
	_ = d.Set("policy_urn", attachment.PolicyURN)
	_ = d.Set("project_id", projectID)
	_ = d.Set("create_time", timestampToString(attachment.AttachedAt))
	return nil
}

func resourceUCloudIAMGroupPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewDetachPoliciesFromGroupRequest()
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
		if _, err := client.iamconn.DetachPoliciesFromGroup(req); err != nil {
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

func buildUCloudIAMUserPolicyAttachmentID(userName, policyURN, projectID string) string {
	if projectID == "" {
		return fmt.Sprintf(accountScopeAttachmentPrefix+"%s/%s", userName, policyURN)
	}
	return fmt.Sprintf(projectScopeAttachmentPrefix+"%s/%s/%s", projectID, userName, policyURN)
}

func extractUCloudIAMUserPolicyAttachmentID(id string) (userName, policyURN, projectID string, err error) {
	return extractPolicyAttachmentID(id)
}

func buildUCloudIAMGroupPolicyAttachmentID(groupName, policyURN, projectID string) string {
	if projectID == "" {
		return fmt.Sprintf(accountScopeAttachmentPrefix+"%s/%s", groupName, policyURN)
	}
	return fmt.Sprintf(projectScopeAttachmentPrefix+"%s/%s/%s", projectID, groupName, policyURN)
}

func extractUCloudIAMGroupPolicyAttachmentID(id string) (groupName, policyURN, projectID string, err error) {
	groupName, policyURN, projectID, err = extractPolicyAttachmentID(id)
	return groupName, policyURN, projectID, err
}

func extractPolicyAttachmentID(id string) (first, policyURN, projectID string, err error) {
	if strings.HasPrefix(id, accountScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, accountScopeAttachmentPrefix), "/", 2)
		if len(items) != 2 || items[0] == "" || items[1] == "" {
			return "", "", "", errors.New("fail to parse id")
		}
		return items[0], items[1], "", nil
	}
	if strings.HasPrefix(id, projectScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, projectScopeAttachmentPrefix), "/", 3)
		if len(items) != 3 || items[0] == "" || items[1] == "" || items[2] == "" {
			return "", "", "", errors.New("fail to parse id")
		}
		return items[1], items[2], items[0], nil
	}
	return "", "", "", errors.New("fail to parse id")
}
