package label

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudLabels() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLabelsRead,
		Schema: map[string]*schema.Schema{
			"key_regex": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"projects": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"id": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"name": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"resource_types": {
									Type:     schema.TypeList,
									Computed: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"disabled_resource_types": {
									Type:     schema.TypeList,
									Computed: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							}},
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudLabelsRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading label list, %s", err)
	}
	limit := 100
	offset := 0
	var labels []labelapi.ListLabelsLabel
	for {
		req := client.NewListLabelsRequest()
		req.Category = ucloud.String(CustomLabelCategory)
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.ListLabels(req)
		if err != nil {
			return fmt.Errorf("error on reading label list, %s", err)
		}

		for _, item := range resp.Labels {
			if keyRegex, ok := data.GetOk("key_regex"); ok {
				matched, err := regexp.Match(keyRegex.(string), []byte(item.Key))
				if err != nil {
					return fmt.Errorf("error on matching key regex, %s", err)
				}
				if !matched {
					continue
				}
			}
			labels = append(labels, item)
		}
		if len(resp.Labels) < limit {
			break
		}
		offset += limit
	}

	labelsData := make([]map[string]interface{}, 0)
	ids := make([]string, 0)
	for _, item := range labels {
		ids = append(ids, buildUCloudLabelID(item.Key, item.Value))
		listProjectsReq := client.NewListProjectsByLabelsRequest()
		listProjectsReq.Labels = []labelapi.ListProjectsByLabelsParamLabels{{Key: &item.Key, Value: &item.Value}}
		resp, err := client.ListProjectsByLabels(listProjectsReq)
		if err != nil {
			return fmt.Errorf("error on listing projects by labels, %s", err)
		}
		projectsData := make([]map[string]interface{}, 0)
		for _, project := range resp.Projects {
			projectsData = append(projectsData, map[string]interface{}{
				"id":                      project.ProjectId,
				"name":                    project.ProjectName,
				"resource_types":          project.ResourceTypes,
				"disabled_resource_types": project.DisabledResourceTypes,
			})
		}

		labelsData = append(labelsData, map[string]interface{}{
			"key":      item.Key,
			"value":    item.Value,
			"projects": projectsData,
		})
	}
	data.SetId(hashStringArray(ids))
	data.Set("total_count", len(labelsData))
	if err := data.Set("labels", labelsData); err != nil {
		return err
	}

	if outputFile, ok := data.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), labelsData)
	}
	return nil
}
