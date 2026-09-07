package iam

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func dataSourceUCloudIAMUsers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudIAMUsersRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"login_enable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudIAMUsersRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewListUsersRequest()
	limit := 100
	offset := 0
	var result []iam.Users
	for {
		resp, err := client.iamconn.ListUsers(req)
		if err != nil {
			return fmt.Errorf("error on reading user list, %s", err)
		}
		if len(resp.Users) < 1 {
			break
		}
		result = append(result, resp.Users...)
		if len(resp.Users) < limit {
			break
		}
		offset += limit
		_ = offset
	}

	users := result
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		users = make([]iam.Users, 0, len(result))
		for _, user := range result {
			if r != nil && !r.MatchString(user.UserName) {
				continue
			}
			users = append(users, user)
		}
	}

	if group, ok := d.GetOk("group_name"); ok {
		usersForGroup, err := client.describeGroupMembership(group.(string))
		if err != nil {
			return fmt.Errorf("error on reading user list when get membership, %s", err)
		}
		userMap := make(map[string]struct{})
		for _, user := range usersForGroup {
			userMap[user.UserName] = struct{}{}
		}
		filtered := make([]iam.Users, 0, len(users))
		for _, user := range users {
			if _, ok := userMap[user.UserName]; ok {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	ids := make([]string, 0, len(users))
	data := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		loginProfile, err := client.describeLoginProfile(user.UserName)
		if err != nil {
			return fmt.Errorf("error on reading user list when get login profile, %s", err)
		}
		ids = append(ids, user.UserName)
		data = append(data, map[string]interface{}{
			"name":         user.UserName,
			"display_name": user.DisplayName,
			"email":        user.Email,
			"status":       user.Status,
			"login_enable": loginProfile.Status == iamStatusActive,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("users", data); err != nil {
		return err
	}
	if err := d.Set("names", ids); err != nil {
		return err
	}
	return nil
}
