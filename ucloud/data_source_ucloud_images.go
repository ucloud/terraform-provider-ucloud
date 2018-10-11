package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudImages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudImagesRead,
		Schema: map[string]*schema.Schema{
			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"image_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"os_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"size": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"availability_zone": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"os_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"os_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"features": &schema.Schema{
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},

						"create_time": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudImagesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uhostconn

	req := conn.NewDescribeImageRequest()

	if v, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("image_type"); ok {
		req.ImageType = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("os_type"); ok {
		req.OsType = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(v.(string))
	}

	var images []uhost.UHostImageSet
	var limit int = 100
	var totalCount int
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeImage(req)
		if err != nil {
			return fmt.Errorf("error in read image list, %s", err)
		}

		if resp == nil || len(resp.ImageSet) < 1 {
			break
		}

		images = append(images, resp.ImageSet...)

		totalCount = totalCount + resp.TotalCount

		if len(resp.ImageSet) < limit {
			break
		}

		offset = offset + limit
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudImagesSave(d, images)
	if err != nil {
		return fmt.Errorf("error in read image list, %s", err)
	}

	return nil
}

func dataSourceUCloudImagesSave(d *schema.ResourceData, projects []uhost.UHostImageSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range projects {
		ids = append(ids, item.ImageId)
		data = append(data, map[string]interface{}{
			"id":                item.ImageId,
			"name":              item.ImageName,
			"type":              item.ImageType,
			"availability_zone": item.Zone,
			"os_type":           item.OsType,
			"os_name":           item.OsName,
			"features":          item.Features,
			"create_time":       timestampToString(item.CreateTime),
			"size":              item.ImageSize,
			"description":       item.ImageDescription,
			"status":            item.State,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("images", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
