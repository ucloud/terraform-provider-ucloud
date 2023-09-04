package ucloud

import (
	"fmt"
	"regexp"

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
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
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

						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"os_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"os_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
		req.ImageType = ucloud.String(upperCamelCvt.convert(v.(string)))
	}

	resp, err := client.getBareMetalImages(*req)
	if err != nil {
		return fmt.Errorf("error on reading bare metal images, %s", err)
	}
	searchedIds, searchedIdsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	osType, osTypeOk := d.GetOk("os_type")
	instanceType, instanceTypeOk := d.GetOk("instance_type")
	ids := make([]string, 0)
	images := make([]map[string]interface{}, 0)
	var r *regexp.Regexp
	if nameRegex != "" {
		r = regexp.MustCompile(nameRegex.(string))
	}
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
		if nameRegexOk && r != nil && !r.MatchString(item.ImageName) {
			continue
		}
		if searchedIdsOk && !isStringIn(item.ImageId, schemaSetToStringSlice(searchedIds)) {
			continue
		}
		image := map[string]interface{}{
			"id":                item.ImageId,
			"name":              item.ImageName,
			"availability_zone": req.Zone,
			"type":              upperCamelCvt.convert(item.ImageType),
			"os_type":           upperCamelCvt.convert(item.OsType),
			"os_name":           item.OsName,
			"size":              item.ImageSize,
			"description":       item.ImageDescription,
			"status":            item.State,
		}
		ids = append(ids, item.ImageId)
		images = append(images, image)
	}

	d.Set("images", images)
	d.Set("total_count", len(images))
	d.SetId(hashStringArray(ids))
	d.Set("ids", ids)
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), images)
	}
	return nil
}
