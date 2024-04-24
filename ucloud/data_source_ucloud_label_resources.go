package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudLabelResources() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLabelResourcesRead,
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources": {
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudLabelResourcesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	limit := 100
	offset := 0
	resourcesData := make([]map[string]interface{}, 0)
	ids := make([]string, 0)
	for {
		listResourcesReq := conn.NewListResourcesByLabelsRequest()
		listResourcesReq.ResourceTypes = interfaceSliceToStringSlice(d.Get("resource_types").([]interface{}))
		listResourcesReq.ProjectIds = interfaceSliceToStringSlice(d.Get("project_ids").([]interface{}))
		listResourcesReq.Limit = ucloud.Int(limit)
		listResourcesReq.Offset = ucloud.Int(offset)
		listResourcesReq.Labels = []label.ListResourcesByLabelsParamLabels{
			{
				Key:   ucloud.String(d.Get("key").(string)),
				Value: ucloud.String(d.Get("value").(string)),
			},
		}
		listResourcesResp, err := conn.ListResourcesByLabels(listResourcesReq)
		if err != nil {
			return fmt.Errorf("error on reading resources list, %s", err)
		}

		if len(listResourcesResp.Resources) < 1 {
			break
		}
		for _, resource := range listResourcesResp.Resources {
			ids = append(ids, resource.ResourceId)
			resourcesData = append(resourcesData, map[string]interface{}{
				"id":   resource.ResourceId,
				"name": resource.ResourceName,
				"type": resource.ResourceType,
			})
		}
		if len(listResourcesResp.Resources) < limit {
			break
		}
		offset = offset + limit
	}
	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(resourcesData))
	if err := d.Set("resources", resourcesData); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), resourcesData)
	}
	return nil
}
