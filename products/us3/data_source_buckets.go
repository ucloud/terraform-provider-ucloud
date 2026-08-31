package us3

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
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
						"source_domain_names": {
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

func dataSourceUCloudUS3BucketsRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading us3 bucket list, %s", err)
	}

	var allBuckets []ufile.UFileBucketSet
	const limit = 100
	offset := 0
	for {
		req := client.NewDescribeBucketRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := client.DescribeBucket(req)
		if err != nil {
			return fmt.Errorf("error on reading us3 bucket list, %s", err)
		}
		if resp == nil || len(resp.DataSet) < 1 {
			break
		}
		allBuckets = append(allBuckets, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
		offset += limit
	}

	buckets := allBuckets
	if nameRegex, ok := data.GetOk("name_regex"); ok {
		var matcher *regexp.Regexp
		if nameRegex != "" {
			matcher = regexp.MustCompile(nameRegex.(string))
		}
		buckets = nil
		for _, bucket := range allBuckets {
			if matcher != nil && !matcher.MatchString(bucket.BucketName) {
				continue
			}
			buckets = append(buckets, bucket)
		}
	}

	details := make([]ufile.UFileBucketSet, 0, len(buckets))
	for _, bucket := range buckets {
		instance, err := describeUS3BucketById(client, bucket.BucketName)
		if err != nil {
			return fmt.Errorf("error on reading us3 bucket %q, %s", bucket.BucketName, err)
		}
		details = append(details, *instance)
	}
	if err := dataSourceUCloudUS3BucketsSave(data, details); err != nil {
		return fmt.Errorf("error on reading us3 bucket list, %s", err)
	}
	return nil
}

func dataSourceUCloudUS3BucketsSave(data *schema.ResourceData, buckets []ufile.UFileBucketSet) error {
	ids := make([]string, 0, len(buckets))
	result := make([]map[string]interface{}, 0, len(buckets))
	for _, bucket := range buckets {
		ids = append(ids, bucket.BucketName)
		result = append(result, map[string]interface{}{
			"type":                bucket.Type,
			"name":                bucket.BucketName,
			"tag":                 bucket.Tag,
			"create_time":         timestampToString(bucket.CreateTime),
			"source_domain_names": bucket.Domain.Src,
		})
	}

	data.SetId(hashStringArray(ids))
	data.Set("total_count", len(result))
	if err := data.Set("us3_buckets", result); err != nil {
		return err
	}
	if outputFile, ok := data.GetOk("output_file"); ok && outputFile.(string) != "" {
		if err := writeToFile(outputFile.(string), result); err != nil {
			return err
		}
	}
	return nil
}
