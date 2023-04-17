package ucloud

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMUserCreate,
		Update: resourceUCloudIAMUserUpdate,
		Read:   resourceUCloudIAMUserRead,
		Delete: resourceUCloudIAMUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"is_frozen": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"login_enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIAMUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewCreateUserRequest()
	req.AccessKeyStatus = ucloud.String(iamStatusInactive)
	req.UserName = ucloud.String(d.Get("name").(string))
	if val, ok := d.GetOk("display_name"); ok {
		req.DisplayName = ucloud.String(val.(string))
	}
	if d.Get("login_enable").(bool) {
		if val, ok := d.GetOk("email"); ok {
			req.Email = ucloud.String(val.(string))
		} else {
			return fmt.Errorf("%q is required when %q is true", "email", "login_enable")
		}
		req.LoginProfileStatus = ucloud.String(iamStatusActive)
	} else {
		if _, ok := d.GetOk("email"); ok {
			return fmt.Errorf("the system will generate a random %q when %q is true", "email", "login_enable")
		}

		req.LoginProfileStatus = ucloud.String(iamStatusInactive)
	}
	resp, err := conn.CreateUser(req)
	if err != nil {
		if resp != nil && resp.RetCode == 11208 {
			return fmt.Errorf("error on creating user because user name already exists, %s", err)
		}
		return fmt.Errorf("error on creating user, %s", err)
	}
	d.SetId(d.Get("name").(string))

	if d.Get("is_frozen").(bool) {
		updateReq := conn.NewUpdateUserRequest()
		updateReq.UserName = ucloud.String(d.Get("name").(string))
		updateReq.Status = ucloud.String(iamStatusFrozen)
		_, err = conn.UpdateUser(updateReq)
		if err != nil {
			return fmt.Errorf("error on creating user when set frozen, %s", err)
		}
	}
	return resourceUCloudIAMUserRead(d, meta)
}

func resourceUCloudIAMUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	needUpdateLoginProfile := false
	profileReq := conn.NewUpdateLoginProfileRequest()
	profileReq.UserName = ucloud.String(d.Id())
	if d.HasChange("email") {
		old, _ := d.GetChange("email")
		if !strings.HasSuffix(old.(string), ".apiuser.ucloud.cn") {
			return fmt.Errorf("if you want to change email, please manage it through the console or API")
		}
		profileReq.UserEmail = ucloud.String(d.Get("email").(string))
		needUpdateLoginProfile = true
	}
	if d.HasChange("login_enable") {
		if !d.HasChange("email") && strings.HasSuffix(d.Get("email").(string), ".apiuser.ucloud.cn") {

		}
		status := iamStatusInactive
		if d.Get("login_enable").(bool) {
			status = iamStatusActive
		}
		profileReq.Status = ucloud.String(status)
		needUpdateLoginProfile = true
	}
	if needUpdateLoginProfile {
		_, err := conn.UpdateLoginProfile(profileReq)
		if err != nil {
			return fmt.Errorf("error on update user when update login profile, %s", err)
		}
	}

	if d.HasChanges("display_name", "is_frozen") {
		req := conn.NewUpdateUserRequest()
		req.UserName = ucloud.String(d.Id())
		if d.HasChange("display_name") {
			displayName := d.Get("display_name").(string)
			if displayName == "" {
				return errors.New("display_name cannot be updated to empty string")
			}
			req.DisplayName = ucloud.String(displayName)
		}
		if d.HasChange("is_frozen") {
			if d.Get("is_frozen").(bool) {
				req.Status = ucloud.String(iamStatusFrozen)
			} else {
				req.Status = ucloud.String(iamStatusActive)
			}
		}
		_, err := conn.UpdateUser(req)
		if err != nil {
			return fmt.Errorf("error on update user, %s", err)
		}
	}
	return resourceUCloudIAMUserRead(d, meta)
}

func resourceUCloudIAMUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	user, err := client.describeUser(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading user %q, %s", d.Id(), err)
	}
	loginProfile, err := client.describeLoginProfile(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading login profile %q, %s", d.Id(), err)
	}
	d.Set("name", user.UserName)
	d.Set("display_name", user.DisplayName)
	d.Set("email", user.Email)
	d.Set("status", user.Status)
	d.Set("is_frozen", user.Status == iamStatusFrozen)
	d.Set("login_enable", loginProfile.Status == iamStatusActive)
	return nil
}

func resourceUCloudIAMUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewDeleteUserRequest()
	req.UserName = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteUser(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting user %q, %s", d.Id(), err))
		}

		_, err := client.describeUser(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading user when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified user %q has not been deleted due to unknown error", d.Id()))
	})
}
