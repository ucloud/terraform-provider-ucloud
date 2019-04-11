package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
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
	client := meta.(*UCloudClient)
	conn := client.uaccountconn

	req := conn.NewGetRegionRequest()

	resp, err := conn.GetRegion(req)
	if err != nil {
		return fmt.Errorf("error on reading region list, %s", err)
	}

	var zones []uaccount.RegionInfo
	for _, item := range resp.Regions {
		if item.Region == client.region {
			zones = append(zones, item)
		}
	}

	err = dataSourceUCloudZonesSave(d, zones, meta)
	if err != nil {
		return fmt.Errorf("error on reading region list, %s", err)
	}

	return nil
}

func dataSourceUCloudZonesSave(d *schema.ResourceData, zones []uaccount.RegionInfo, meta interface{}) error {
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
