package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
						"comments": {
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
	conn := meta.(*UCloudClient).iamconn

	req := conn.NewListGroupsRequest()

	limit := 100
	offset := 0
	var res []iam.Group
	for {
		resp, err := conn.ListGroups(req)
		if err != nil {
			return fmt.Errorf("error on reading group list, %s", err)
		}
		if len(resp.Groups) < 1 {
			break
		}
		res = append(res, resp.Groups...)
		if len(resp.Groups) < limit {
			break
		}
		offset = offset + limit
	}
	var groups []iam.Group
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range res {
			if r != nil && !r.MatchString(v.GroupName) {
				continue
			}

			groups = append(groups, v)
		}
	} else {
		groups = res
	}

	ids := []string{}
	data := []map[string]interface{}{}

	for _, g := range groups {
		ids = append(ids, g.GroupName)
		data = append(data, map[string]interface{}{
			"name":     g.GroupName,
			"comments": g.Description,
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
