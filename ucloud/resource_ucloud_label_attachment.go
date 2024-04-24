package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLabelAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLabelAttachmentCreate,
		Update: nil,
		Read:   resourceUCloudLabelAttachmentRead,
		Delete: resourceUCloudLabelAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUCloudLabelAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	req := conn.NewBindLabelsRequest()
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	resource := d.Get("resource").(string)
	label := label.BindLabelsParamLabels{
		Key:   &key,
		Value: &value,
	}
	req.Labels = append(req.Labels, label)
	req.ResourceIds = []string{resource}
	_, err := conn.BindLabels(req)
	if err != nil {
		return fmt.Errorf("error on binding label, %s", err)
	}
	d.SetId(buildUCloudLabelAttachmentID(key, value, resource))
	return resourceUCloudLabelAttachmentRead(d, meta)
}
func resourceUCloudLabelAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	key, value, resource, err := parseUCloudLabelAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing label id, %s", err)
	}
	client := meta.(*UCloudClient)
	_, err = client.describeLabelAttachment(key, value, resource)
	if err != nil {
		return fmt.Errorf("error on describing label, %s", err)
	}
	d.Set("key", key)
	d.Set("value", value)
	d.Set("resource", resource)
	return nil
}

func resourceUCloudLabelAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	req := conn.NewUnbindLabelsRequest()
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	req.Labels = append(req.Labels, label.UnbindLabelsParamLabels{
		Key:   ucloud.String(key),
		Value: ucloud.String(value),
	})
	req.ResourceIds = []string{d.Get("resource").(string)}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.UnbindLabels(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on unbind label %q, %s", d.Id(), err))
		}

		_, err := client.describeLabelAttachment(key, value, d.Get("resource").(string))
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading label when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified label attachment %q has not been deleted due to unknown error", d.Id()))
	})
}

const UCloudLabelAttachmentIDSeperator = "#"

func buildUCloudLabelAttachmentID(key, value, resourceId string) string {
	return strings.Join([]string{key, value, resourceId}, UCloudLabelAttachmentIDSeperator)
}

func parseUCloudLabelAttachmentID(id string) (key string, value string, resourceId string, err error) {
	items := strings.Split(id, UCloudLabelAttachmentIDSeperator)
	if len(items) != 3 {
		return "", "", "", fmt.Errorf("invalid label attachment id: %s", id)
	}
	return items[0], items[1], items[2], nil
}
