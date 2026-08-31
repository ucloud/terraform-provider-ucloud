package label

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
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

func resourceUCloudLabelCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating label, %s", err)
	}
	req := client.NewCreateLabelsRequest()
	key := data.Get("key").(string)
	value := data.Get("value").(string)
	req.Labels = append(req.Labels, labelapi.CreateLabelsParamLabels{
		Key:   &key,
		Value: &value,
	})
	if _, err := client.CreateLabels(req); err != nil {
		return fmt.Errorf("error on creating label, %s", err)
	}
	data.SetId(buildUCloudLabelID(key, value))
	return resourceUCloudLabelRead(data, meta)
}

func resourceUCloudLabelRead(data *schema.ResourceData, meta interface{}) error {
	key, value, err := parseUCloudLabelID(data.Id())
	if err != nil {
		return fmt.Errorf("error on parsing label id, %s", err)
	}
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading label, %s", err)
	}
	labelInfo, err := describeLabel(client, key, value)
	if err != nil {
		return fmt.Errorf("error on describing label, %s", err)
	}
	data.Set("key", labelInfo.Key)
	data.Set("value", labelInfo.Value)
	return nil
}

func resourceUCloudLabelDelete(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting label, %s", err)
	}
	req := client.NewDeleteLabelsRequest()
	key := data.Get("key").(string)
	value := data.Get("value").(string)
	req.Labels = append(req.Labels, labelapi.DeleteLabelsParamLabels{
		Key:   ucloud.String(key),
		Value: ucloud.String(value),
	})

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.DeleteLabels(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting label %q, %s", data.Id(), err))
		}

		_, err := describeLabel(client, key, value)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading label when deleting %q, %s", data.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified label %q has not been deleted due to unknown error", data.Id()))
	})
}
