package ucloud

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
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewListUsersRequest()

	limit := 100
	offset := 0
	var res []iam.Users
	for {
		resp, err := conn.ListUsers(req)
		if err != nil {
			return fmt.Errorf("error on reading user list, %s", err)
		}
		if len(resp.Users) < 1 {
			break
		}
		res = append(res, resp.Users...)
		if len(resp.Users) < limit {
			break
		}
		offset = offset + limit
	}
	var users []iam.Users
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range res {
			if r != nil && !r.MatchString(v.UserName) {
				continue
			}
			users = append(users, v)
		}
	} else {
		users = res
	}

	if group, ok := d.GetOk("group_name"); ok {
		userForGroup, err := client.describeGroupMembership(group.(string))
		if err != nil {
			return fmt.Errorf("error on reading user list when get membership, %s", err)
		}
		userMap := make(map[string]struct{}, 0)
		for _, v := range userForGroup {
			userMap[v.UserName] = struct{}{}
		}
		var us []iam.Users
		for _, v := range users {
			if _, ok := userMap[v.UserName]; ok {
				us = append(us, v)
			}
		}
		users = us
	}

	ids := []string{}
	data := []map[string]interface{}{}
	for _, u := range users {
		loginProfile, err := client.describeLoginProfile(u.UserName)
		if err != nil {
			return fmt.Errorf("error on reading user list when get login profile %q, %s", d.Id(), err)
		}
		ids = append(ids, u.UserName)
		data = append(data, map[string]interface{}{
			"name":         u.UserName,
			"display_name": u.DisplayName,
			"email":        u.Email,
			"status":       u.Status,
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
