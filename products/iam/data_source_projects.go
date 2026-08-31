package iam

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	projects, err := client.listIAMProject()
	if err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}
	filteredProjects := make([]iam.Project, 0, len(projects))
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, project := range projects {
			if r != nil && !r.MatchString(project.ProjectName) {
				continue
			}
			filteredProjects = append(filteredProjects, project)
		}
	} else {
		filteredProjects = projects
	}

	if err := dataSourceUCloudIAMProjectsSave(d, filteredProjects); err != nil {
		return fmt.Errorf("error on reading project list, %s", err)
	}
	return nil
}

func dataSourceUCloudIAMProjectsSave(d *schema.ResourceData, projects []iam.Project) error {
	ids := make([]string, 0, len(projects))
	data := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		ids = append(ids, project.ProjectID)
		data = append(data, map[string]interface{}{
			"id":          project.ProjectID,
			"name":        project.ProjectName,
			"user_count":  project.UserCount,
			"create_time": timestampToString(project.CreatedAt),
		})
	}

	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	if err := d.Set("projects", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		_ = writeToFile(outputFile.(string), data)
	}
	return nil
}
