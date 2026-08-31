package label

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
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

func resourceUCloudLabelAttachmentCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating label attachment, %s", err)
	}
	req := client.NewBindLabelsRequest()
	key := data.Get("key").(string)
	value := data.Get("value").(string)
	resourceID := data.Get("resource").(string)
	req.Labels = append(req.Labels, labelapi.BindLabelsParamLabels{
		Key:   &key,
		Value: &value,
	})
	req.ResourceIds = []string{resourceID}
	if _, err := client.BindLabels(req); err != nil {
		return fmt.Errorf("error on binding label, %s", err)
	}
	data.SetId(buildUCloudLabelAttachmentID(key, value, resourceID))
	return resourceUCloudLabelAttachmentRead(data, meta)
}

func resourceUCloudLabelAttachmentRead(data *schema.ResourceData, meta interface{}) error {
	key, value, resourceID, err := parseUCloudLabelAttachmentID(data.Id())
	if err != nil {
		return fmt.Errorf("error on parsing label id, %s", err)
	}
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading label attachment, %s", err)
	}
	if _, err = describeLabelAttachment(client, key, value, resourceID); err != nil {
		return fmt.Errorf("error on describing label, %s", err)
	}
	data.Set("key", key)
	data.Set("value", value)
	data.Set("resource", resourceID)
	return nil
}

func resourceUCloudLabelAttachmentDelete(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting label attachment, %s", err)
	}
	req := client.NewUnbindLabelsRequest()
	key := data.Get("key").(string)
	value := data.Get("value").(string)
	resourceID := data.Get("resource").(string)
	req.Labels = append(req.Labels, labelapi.UnbindLabelsParamLabels{
		Key:   ucloud.String(key),
		Value: ucloud.String(value),
	})
	req.ResourceIds = []string{resourceID}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.UnbindLabels(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on unbind label %q, %s", data.Id(), err))
		}

		_, err := describeLabelAttachment(client, key, value, resourceID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading label when deleting %q, %s", data.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified label attachment %q has not been deleted due to unknown error", data.Id()))
	})
}
