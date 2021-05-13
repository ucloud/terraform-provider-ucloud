package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudUS3Buckets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUS3BucketsRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"us3_buckets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"src_domain_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudUS3BucketsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).us3conn
	var allUS3Buckets []ufile.UFileBucketSet
	var us3Buckets []ufile.UFileBucketSet
	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeBucketRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeBucket(req)
		if err != nil {
			return fmt.Errorf("error on reading ufs list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allUS3Buckets = append(allUS3Buckets, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	//ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allUS3Buckets {
			if r != nil && !r.MatchString(v.BucketName) {
				continue
			}

			//if idsOk && !isStringIn(v.VolumeId, schemaSetToStringSlice(ids)) {
			//	continue
			//}
			us3Buckets = append(us3Buckets, v)
		}
	} else {
		us3Buckets = allUS3Buckets
	}

	err := dataSourceUCloudUS3BucketsSave(d, us3Buckets)
	if err != nil {
		return fmt.Errorf("error on reading ufs list, %s", err)
	}

	return nil
}

func dataSourceUCloudUS3BucketsSave(d *schema.ResourceData, us3Buckets []ufile.UFileBucketSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range us3Buckets {
		ids = append(ids, item.BucketName)

		data = append(data, map[string]interface{}{
			"type":             item.Type,
			"name":             item.BucketName,
			"tag":              item.Tag,
			"create_time":      timestampToString(item.CreateTime),
			"src_domain_names": item.Domain.Src,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	//d.Set("ids", ids)
	if err := d.Set("us3_buckets", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		if err := writeToFile(outputFile.(string), data); err != nil {
			return err
		}
	}

	return nil
}
