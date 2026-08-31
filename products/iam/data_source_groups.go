package iam

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func dataSourceUCloudIAMGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudIAMGroupsRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudIAMGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	req := client.iamconn.NewListGroupsRequest()
	limit := 100
	offset := 0
	var result []iam.Group
	for {
		resp, err := client.iamconn.ListGroups(req)
		if err != nil {
			return fmt.Errorf("error on reading group list, %s", err)
		}
		if len(resp.Groups) < 1 {
			break
		}
		result = append(result, resp.Groups...)
		if len(resp.Groups) < limit {
			break
		}
		offset += limit
		_ = offset
	}

	groups := result
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		groups = make([]iam.Group, 0, len(result))
		for _, group := range result {
			if r != nil && !r.MatchString(group.GroupName) {
				continue
			}
			groups = append(groups, group)
		}
	}

	ids := make([]string, 0, len(groups))
	data := make([]map[string]interface{}, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.GroupName)
		data = append(data, map[string]interface{}{
			"name":    group.GroupName,
			"comment": group.Description,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("groups", data); err != nil {
		return err
	}
	if err := d.Set("names", ids); err != nil {
		return err
	}
	return nil
}
