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

func resourceUCloudLabel() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLabelCreate,
		Update: nil,
		Read:   resourceUCloudLabelRead,
		Delete: resourceUCloudLabelDelete,
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
		},
	}
}

func resourceUCloudLabelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	req := conn.NewCreateLabelsRequest()
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	label := label.CreateLabelsParamLabels{
		Key:   &key,
		Value: &value,
	}
	req.Labels = append(req.Labels, label)
	_, err := conn.CreateLabels(req)
	if err != nil {
		return fmt.Errorf("error on creating label, %s", err)
	}
	d.SetId(buildUCloudLabelID(key, value))
	return resourceUCloudLabelRead(d, meta)
}
func resourceUCloudLabelRead(d *schema.ResourceData, meta interface{}) error {
	key, value, err := parseUCloudLabelID(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing label id, %s", err)
	}
	client := meta.(*UCloudClient)
	label, err := client.describeLabel(key, value)
	if err != nil {
		return fmt.Errorf("error on describing label, %s", err)
	}
	d.Set("key", label.Key)
	d.Set("value", label.Value)
	return nil
}

func resourceUCloudLabelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.labelconn
	req := conn.NewDeleteLabelsRequest()
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	req.Labels = append(req.Labels, label.DeleteLabelsParamLabels{
		Key:   ucloud.String(key),
		Value: ucloud.String(value),
	})

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteLabels(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting label %q, %s", d.Id(), err))
		}

		_, err := client.describeLabel(key, value)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading label when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified label %q has not been deleted due to unknown error", d.Id()))
	})
}

const UCloudLabelIDSeperator = "#"

func buildUCloudLabelID(key, value string) string {
	return key + UCloudLabelIDSeperator + value
}

func parseUCloudLabelID(id string) (string, string, error) {
	items := strings.Split(id, UCloudLabelIDSeperator)
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid label id: %s", id)
	}
	return items[0], items[1], nil
}
