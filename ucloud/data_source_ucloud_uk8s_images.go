package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudUK8SImages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUK8SImagesRead,
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"host_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "uhost",
				ValidateFunc: validation.StringInSlice([]string{"uhost", "uphost"}, false),
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

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

						"zone_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"not_support_gpu": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudUK8SImagesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uk8sconn

	req := conn.NewDescribeUK8SImageRequest()

	if v, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(v.(string))
	}

	resp, err := conn.DescribeUK8SImage(req)
	if err != nil {
		return fmt.Errorf("error on reading uk8s image list, %s", err)
	}

	var images []uk8s.ImageInfo
	if d.Get("host_type").(string) == "uphost" {
		images = resp.PHostImageSet
	} else {
		images = resp.ImageSet
	}

	var filteredImages []uk8s.ImageInfo
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, image := range images {
			if r.MatchString(image.ImageName) {
				filteredImages = append(filteredImages, image)
			}
		}
	} else {
		filteredImages = images
	}

	return dataSourceUCloudUK8SImagesSave(d, filteredImages)
}

func dataSourceUCloudUK8SImagesSave(d *schema.ResourceData, images []uk8s.ImageInfo) error {
	var ids []string
	var data []map[string]interface{}

	for _, item := range images {
		ids = append(ids, item.ImageId)
		data = append(data, map[string]interface{}{
			"id":              item.ImageId,
			"name":            item.ImageName,
			"zone_id":         item.ZoneId,
			"not_support_gpu": item.NotSupportGPU,
		})
	}

	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	if err := d.Set("images", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
