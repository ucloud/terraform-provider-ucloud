package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceUCloudUHubRepositoryImageTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUHubRepositoryImageTagsRead,
		Schema: map[string]*schema.Schema{
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"image_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"repository_image_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

type tagSet struct {
	// 镜像更新时间
	UpdateTime string `required:"true"`
	// Tag名称
	TagName string `required:"true"`

	Digest string
}

type GetImageTagResponse struct {
	// tag总数
	TotalCount int `required:"true"`
	// tag列表
	TagSet []tagSet `required:"true"`
}

func dataSourceUCloudUHubRepositoryImageTagsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).genericconn

	var allRepositoryImageTags []tagSet
	var offset int
	limit := 100
	for {
		req := conn.NewGenericRequest()
		m := map[string]interface{}{
			"Action":    "GetImageTag",
			"RepoName":  d.Get("repository_name").(string),
			"ImageName": d.Get("image_name").(string),
			"Limit":     limit,
			"Offset":    offset,
		}
		if v, ok := d.GetOk("name"); ok {
			m["TagName"] = v.(string)
		}

		err := req.SetPayload(m)
		if err != nil {
			return fmt.Errorf("error on setting request when reading repository image tag, %s", err)
		}

		r, err := conn.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on reading repository image tag, %s", err)
		}

		var resp GetImageTagResponse
		if err := r.Unmarshal(&resp); err != nil {
			return fmt.Errorf("error on reading repository image tag when parse resp, %s", err)
		}

		if len(resp.TagSet) < 1 {
			break
		}

		for _, v := range resp.TagSet {
			allRepositoryImageTags = append(allRepositoryImageTags, v)
		}

		if len(resp.TagSet) < limit {
			break
		}

		offset = offset + limit
	}

	err := dataSourceUCloudRepositoryImageTagsSave(d, allRepositoryImageTags)
	if err != nil {
		return fmt.Errorf("error on reading repository image tag, %s", err)
	}

	return nil
}

func dataSourceUCloudRepositoryImageTagsSave(d *schema.ResourceData, projects []tagSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range projects {
		ids = append(ids, item.TagName)
		data = append(data, map[string]interface{}{
			"name":   item.TagName,
			"digest": item.Digest,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))

	if err := d.Set("repository_image_tags", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
