package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceUCloudIAMProjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudIAMProjectsRead,
		Schema: map[string]*schema.Schema{
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

						"user_count": {
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

func dataSourceUCloudIAMProjectsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	projects, err := client.listIAMProject()
	if err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}
	filteredProjects := make([]iam.Project, 0)
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range projects {
			if r != nil && !r.MatchString(v.ProjectName) {
				continue
			}

			filteredProjects = append(filteredProjects, v)
		}
	} else {
		filteredProjects = projects
	}

	err = dataSourceUCloudIAMProjectsSave(d, filteredProjects)
	if err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}

	return nil
}

func dataSourceUCloudIAMProjectsSave(d *schema.ResourceData, projects []iam.Project) error {
	var ids []string
	var data []map[string]interface{}

	for _, item := range projects {
		ids = append(ids, item.ProjectID)
		data = append(data, map[string]interface{}{
			"id":          item.ProjectID,
			"name":        item.ProjectName,
			"user_count":  item.UserCount,
			"create_time": timestampToString(item.CreatedAt),
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
