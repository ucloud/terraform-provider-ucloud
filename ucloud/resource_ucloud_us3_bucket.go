package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
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

			//"tag": {
			//	Type:         schema.TypeString,
			//	Optional:     true,
			//	ForceNew:     true,
			//	Default:      defaultTag,
			//	ValidateFunc: validateTag,
			//	StateFunc:    stateFuncTag,
			//},

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
	}
}

func resourceUCloudUS3BucketCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.us3conn
	req := conn.NewCreateBucketRequest()
	req.Type = ucloud.String(d.Get("type").(string))
	req.BucketName = ucloud.String(d.Get("name").(string))

	// if tag is empty string, use default tag
	//if v, ok := d.GetOk("tag"); ok {
	//	req.Tag = ucloud.String(v.(string))
	//} else {
	//	req.Tag = ucloud.String(defaultTag)
	//}
	resp, err := conn.CreateBucket(req)
	if err != nil {
		return fmt.Errorf("error on creating us3 bucket, %s", err)
	}

	d.SetId(resp.BucketName)

	return resourceUCloudUS3BucketRead(d, meta)
}

func resourceUCloudUS3BucketUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.us3conn

	d.Partial(true)

	if d.HasChange("type") && !d.IsNewResource() {
		reqBand := conn.NewUpdateBucketRequest()
		reqBand.BucketName = ucloud.String(d.Get("name").(string))
		reqBand.Type = ucloud.String(d.Get("type").(string))

		_, err := conn.UpdateBucket(reqBand)
		if err != nil {
			return fmt.Errorf("error on %s to us3 bucket %q, %s", "UpdateBucket", d.Id(), err)
		}

		d.SetPartial("type")
	}

	d.Partial(false)

	return resourceUCloudUS3BucketRead(d, meta)
}

func resourceUCloudUS3BucketRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instance, err := client.describeUS3BucketById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading us3 bucket %q, %s", d.Id(), err)
	}

	d.Set("type", instance.Type)
	d.Set("name", instance.BucketName)
	d.Set("tag", instance.Tag)
	d.Set("create_time", timestampToString(instance.CreateTime))
	d.Set("src_domain_names", instance.Domain.Src)

	return nil
}

func resourceUCloudUS3BucketDelete(d *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client := meta.(*UCloudClient)
		conn := client.us3conn

		req := conn.NewDeleteBucketRequest()
		req.BucketName = ucloud.String(d.Id())
		if _, err := conn.DeleteBucket(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting us3 bucket %q, %s", d.Id(), err))
		}

		_, err := client.describeUS3BucketById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading us3 bucket when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified us3 bucket %q has not been deleted due to unknown error", d.Id()))
	})
}
