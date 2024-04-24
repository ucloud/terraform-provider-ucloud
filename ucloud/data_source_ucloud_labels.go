package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/label"
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

func dataSourceUCloudLabelsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	limit := 100
	offset := 0
	var labels []label.ListLabelsLabel
	for {
		listLabelsReq := conn.NewListLabelsRequest()
		listLabelsReq.Category = ucloud.String(CustomLabelCategory)
		listLabelsReq.Limit = ucloud.Int(limit)
		listLabelsReq.Offset = ucloud.Int(offset)
		listLabelsResp, err := conn.ListLabels(listLabelsReq)
		if err != nil {
			return fmt.Errorf("error on reading label list, %s", err)
		}

		for _, item := range listLabelsResp.Labels {
			if keyRegex, ok := d.GetOk("key_regex"); ok {
				if matched, err := regexp.Match(keyRegex.(string), []byte(item.Key)); err != nil {
					return fmt.Errorf("error on matching key regex, %s", err)
				} else if !matched {
					continue
				} else {
					labels = append(labels, item)
				}
			} else {
				labels = append(labels, item)
			}
		}
		if len(listLabelsResp.Labels) < limit {
			break
		}
		offset += limit
	}
	labelsData := make([]map[string]interface{}, 0)
	ids := make([]string, 0)
	for _, item := range labels {
		ids = append(ids, buildUCloudLabelID(item.Key, item.Value))
		listProjectsByLabels := conn.NewListProjectsByLabelsRequest()
		listProjectsByLabels.Labels = []label.ListProjectsByLabelsParamLabels{{Key: &item.Key, Value: &item.Value}}
		resp, err := conn.ListProjectsByLabels(listProjectsByLabels)
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
	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(labelsData))
	if err := d.Set("labels", labelsData); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), labelsData)
	}
	return nil
}
