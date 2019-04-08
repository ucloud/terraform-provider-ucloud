package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudProjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudProjectsRead,
		Schema: map[string]*schema.Schema{
			"is_finance": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"parent_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"parent_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"resource_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"member_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudProjectsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uaccountconn

	req := conn.NewGetProjectListRequest()

	if v, ok := d.GetOk("is_finance"); ok {
		req.IsFinance = ucloud.String(boolLowerCvt.convert(v.(bool)))
	}

	resp, err := conn.GetProjectList(req)
	if err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}

	var projects []uaccount.ProjectListInfo

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range resp.ProjectSet {
			if r != nil && !r.MatchString(v.ProjectName) {
				continue
			}

			projects = append(projects, v)
		}
	} else {
		projects = resp.ProjectSet
	}

	err = dataSourceUCloudProjectsSave(d, projects)
	if err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}

	return nil
}

func dataSourceUCloudProjectsSave(d *schema.ResourceData, projects []uaccount.ProjectListInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range projects {
		ids = append(ids, item.ProjectId)
		data = append(data, map[string]interface{}{
			"id":             item.ProjectId,
			"name":           item.ProjectName,
			"parent_id":      item.ParentId,
			"parent_name":    item.ParentName,
			"resource_count": item.ResourceCount,
			"member_count":   item.MemberCount,
			"create_time":    timestampToString(item.CreateTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("projects", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
