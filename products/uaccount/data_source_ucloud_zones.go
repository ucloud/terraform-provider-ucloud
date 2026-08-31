package uaccount

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	sdkuaccount "github.com/ucloud/ucloud-sdk-go/services/uaccount"
)

func dataSourceUCloudZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudZonesRead,
		Schema: map[string]*schema.Schema{
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudZonesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading region list, %s", err)
	}

	req := client.NewGetRegionRequest()

	resp, err := client.GetRegion(req)
	if err != nil {
		return fmt.Errorf("error on reading region list, %s", err)
	}

	var zones []sdkuaccount.RegionInfo
	for _, item := range resp.Regions {
		if item.Region == client.GetConfig().Region {
			zones = append(zones, item)
		}
	}

	if err := dataSourceUCloudZonesSave(d, zones); err != nil {
		return fmt.Errorf("error on reading region list, %s", err)
	}

	return nil
}

func dataSourceUCloudZonesSave(d *schema.ResourceData, zones []sdkuaccount.RegionInfo) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range zones {
		ids = append(ids, item.Zone)
		data = append(data, map[string]interface{}{
			"id": item.Zone,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("zones", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
