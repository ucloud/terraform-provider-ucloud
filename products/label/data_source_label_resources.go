package label

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
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

func dataSourceUCloudLabelResourcesRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading resources list, %s", err)
	}
	limit := 100
	offset := 0
	resourcesData := make([]map[string]interface{}, 0)
	ids := make([]string, 0)
	for {
		req := client.NewListResourcesByLabelsRequest()
		req.ResourceTypes = interfaceSliceToStringSlice(data.Get("resource_types").([]interface{}))
		req.ProjectIds = interfaceSliceToStringSlice(data.Get("project_ids").([]interface{}))
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		req.Labels = []labelapi.ListResourcesByLabelsParamLabels{
			{
				Key:   ucloud.String(data.Get("key").(string)),
				Value: ucloud.String(data.Get("value").(string)),
			},
		}
		resp, err := client.ListResourcesByLabels(req)
		if err != nil {
			return fmt.Errorf("error on reading resources list, %s", err)
		}

		if len(resp.Resources) < 1 {
			break
		}
		for _, resourceInfo := range resp.Resources {
			ids = append(ids, resourceInfo.ResourceId)
			resourcesData = append(resourcesData, map[string]interface{}{
				"id":   resourceInfo.ResourceId,
				"name": resourceInfo.ResourceName,
				"type": resourceInfo.ResourceType,
			})
		}
		if len(resp.Resources) < limit {
			break
		}
		offset += limit
	}
	data.SetId(hashStringArray(ids))
	data.Set("total_count", len(resourcesData))
	if err := data.Set("resources", resourcesData); err != nil {
		return err
	}

	if outputFile, ok := data.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), resourcesData)
	}
	return nil
}
