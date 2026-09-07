package us3

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUS3Bucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUS3BucketCreate,
		Read:   resourceUCloudUS3BucketRead,
		Update: resourceUCloudUS3BucketUpdate,
		Delete: resourceUCloudUS3BucketDelete,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"public",
					"private",
				}, false),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validateUS3BucketName1,
					validateUS3BucketName2,
					validateUS3BucketName3,
					validateUS3BucketName4,
				),
			},
			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
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
	}
}

func resourceUCloudUS3BucketCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating us3 bucket, %s", err)
	}

	tag := defaultTag
	if value, ok := data.GetOk("tag"); ok {
		tag = value.(string)
	}
	req := client.NewGenericRequest()
	if err := req.SetPayload(map[string]interface{}{
		"Action":     "CreateBucket",
		"Type":       data.Get("type").(string),
		"BucketName": data.Get("name").(string),
		"Tag":        tag,
	}); err != nil {
		return fmt.Errorf("error on setting request when creating us3 bucket, %s", err)
	}

	resp, err := client.GenericInvoke(req)
	if err != nil {
		return fmt.Errorf("error on creating us3 bucket, %s", err)
	}
	var result struct {
		BucketName string
	}
	if err := resp.Unmarshal(&result); err != nil {
		return fmt.Errorf("error on creating us3 bucket when parse resp, %s", err)
	}

	data.SetId(result.BucketName)
	return resourceUCloudUS3BucketRead(data, meta)
}

func resourceUCloudUS3BucketUpdate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating us3 bucket, %s", err)
	}

	data.Partial(true)
	if data.HasChange("type") && !data.IsNewResource() {
		req := client.NewUpdateBucketRequest()
		req.BucketName = ucloud.String(data.Get("name").(string))
		req.Type = ucloud.String(data.Get("type").(string))

		if _, err := client.UpdateBucket(req); err != nil {
			return fmt.Errorf("error on %s to us3 bucket %q, %s", "UpdateBucket", data.Id(), err)
		}
		data.SetPartial("type")
	}
	data.Partial(false)

	return resourceUCloudUS3BucketRead(data, meta)
}

func resourceUCloudUS3BucketRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading us3 bucket, %s", err)
	}
	instance, err := describeUS3BucketById(client, data.Id())
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading us3 bucket %q, %s", data.Id(), err)
	}

	data.Set("type", instance.Type)
	data.Set("name", instance.BucketName)
	data.Set("tag", instance.Tag)
	data.Set("create_time", timestampToString(instance.CreateTime))
	data.Set("source_domain_names", instance.Domain.Src)
	return nil
}

func resourceUCloudUS3BucketDelete(data *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client, err := clientFromMeta(meta)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on getting client when deleting us3 bucket, %s", err))
		}

		req := client.NewDeleteBucketRequest()
		req.BucketName = ucloud.String(data.Id())
		if _, err := client.DeleteBucket(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting us3 bucket %q, %s", data.Id(), err))
		}

		_, err = describeUS3BucketById(client, data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading us3 bucket when deleting %q, %s", data.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified us3 bucket %q has not been deleted due to unknown error", data.Id()))
	})
}
