package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"regexp"
)

func dataSourceUCloudUHubRepositoryImages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUHubRepositoryImagesRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"repository_images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latest_tag": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

type repositoryImageSet struct {
	// 镜像名称
	ImageName string `required:"true"`
	// 镜像被下载次数
	PullCount int `required:"true"`
	// 创建时间
	CreateTime string `required:"true"`
	// 修改时间
	UpdateTime string `required:"true"`
	// 最新push的Tag
	LatestTag string `required:"true"`
	// 镜像仓库名称
	RepoName string `required:"true"`
}

type getRepoImageResponse struct {
	TotalCount int `required:"true"`
	// 镜像列表
	ImageSet []repositoryImageSet `required:"true"`
}

func dataSourceUCloudUHubRepositoryImagesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).genericconn

	var allRepositoryImages []repositoryImageSet
	var offset int
	limit := 100
	for {
		req := conn.NewGenericRequest()
		err := req.SetPayload(map[string]interface{}{
			"Action":   "GetRepoImage",
			"RepoName": d.Get("repository_name").(string),
			"Limit":    limit,
			"Offset":   offset,
		})
		if err != nil {
			return fmt.Errorf("error on setting request when reading repository image, %s", err)
		}

		r, err := conn.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on reading repository image, %s", err)
		}

		var resp getRepoImageResponse
		if err := r.Unmarshal(&resp); err != nil {
			return fmt.Errorf("error on reading repository image when parse resp, %s", err)
		}

		if len(resp.ImageSet) < 1 {
			break
		}

		for _, v := range resp.ImageSet {
			allRepositoryImages = append(allRepositoryImages, v)
		}

		if len(resp.ImageSet) < limit {
			break
		}

		offset = offset + limit
	}

	nameRegex, nameRegexOk := d.GetOk("name_regex")

	var filteredRepositoryImages []repositoryImageSet

	if nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, image := range allRepositoryImages {
			if r != nil && !r.MatchString(image.ImageName) {
				continue
			}

			filteredRepositoryImages = append(filteredRepositoryImages, image)
		}
	} else {
		filteredRepositoryImages = allRepositoryImages
	}

	err := dataSourceUCloudRepositoryImagesSave(d, filteredRepositoryImages)
	if err != nil {
		return fmt.Errorf("error on reading repository image, %s", err)
	}

	return nil
}

func dataSourceUCloudRepositoryImagesSave(d *schema.ResourceData, projects []repositoryImageSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range projects {
		ids = append(ids, item.ImageName)
		data = append(data, map[string]interface{}{
			"name":       item.ImageName,
			"latest_tag": item.LatestTag,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))

	if err := d.Set("repository_images", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
