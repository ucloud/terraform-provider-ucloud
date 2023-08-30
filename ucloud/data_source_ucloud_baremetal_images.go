package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudBareMetalImages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudBareMetalImagesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name_regex": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"image_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"os_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: map[string]*schema.Schema{}}, // Define the schema of the "images" here
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceUCloudBareMetalImagesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	req := conn.NewDescribePHostImageRequest()

	// Set the request parameters here based on the schema
	if v, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(v.(string))
	}
	if v, ok := d.GetOk("image_type"); ok {
		req.ImageType = ucloud.String(v.(string))
	}

	instanceType, instanceTypeOk := d.GetOk("instance_type")

	resp, err := conn.DescribePHostImage(req)
	if err != nil {
		return fmt.Errorf("error on reading bare metal images, %s", err)
	}

	osType, osTypeOk := d.GetOk("os_type")

	images := make([]map[string]interface{}, 0)
	for _, item := range resp.ImageSet {
		// Filter images based on os_type
		if osTypeOk && osType.(string) != item.OsType {
			continue
		}

		// Filter images based on instance_type
		if instanceTypeOk {
			supported := false
			for _, support := range item.Support {
				if support == instanceType.(string) {
					supported = true
					break
				}
			}
			if !supported {
				continue
			}
		}

		image := map[string]interface{}{
			"availability_zone": req.Zone,
			"description":       item.ImageDescription,
			"id":                item.ImageId,
			"name":              item.ImageName,
			"type":              item.ImageType,
			"os_name":           item.OsName,
			"os_type":           item.OsType,
		}
		images = append(images, image)
	}

	d.Set("images", images)
	d.Set("total_count", len(images))

	return nil
}
