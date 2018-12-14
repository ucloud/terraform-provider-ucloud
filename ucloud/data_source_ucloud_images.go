package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

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

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"image_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"base", "business", "custom"}, false),
			},

			"os_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"linux", "windows"}, false),
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

	nameRegex, nameRegexOk := d.GetOk("name_regex")

	if v, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("image_type"); ok {
		req.ImageType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	}

	if v, ok := d.GetOk("os_type"); ok {
		req.OsType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	}

	if v, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(v.(string))
	}

	var images []uhost.UHostImageSet
	var totalCount int
	var offset int
	limit := 100
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeImage(req)
		if err != nil {
			return fmt.Errorf("error on reading image list, %s", err)
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

	var filteredImages []uhost.UHostImageSet
	if nameRegexOk {
		r := regexp.MustCompile(nameRegex.(string))
		totalCount = 0
		for _, image := range images {
			if r.MatchString(image.ImageName) && image.State == "Available" {
				filteredImages = append(filteredImages, image)
				totalCount++
			}
		}
	} else {
		filteredImages = images[:]
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudImagesSave(d, filteredImages)
	if err != nil {
		return fmt.Errorf("error on reading image list, %s", err)
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
			"availability_zone": item.Zone,
			"type":              upperCamelCvt.convert(item.ImageType),
			"os_type":           upperCamelCvt.convert(item.OsType),
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
