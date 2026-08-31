package ulb

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBAttachmentCreate,
		Read:   resourceUCloudLBAttachmentRead,
		Update: resourceUCloudLBAttachmentUpdate,
		Delete: resourceUCloudLBAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return isStringIn(old, []string{resourceTypeInstance, lbResourceTypeUHost}) &&
						isStringIn(new, []string{resourceTypeInstance, lbResourceTypeUHost})
				},
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      80,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudLBAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb attachment, %s", err)
	}
	conn := client.ulbconn

	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	resourceID := d.Get("resource_id").(string)
	req := conn.NewAllocateBackendRequest()
	req.ULBId = ucloud.String(lbID)
	req.VServerId = ucloud.String(listenerID)
	req.ResourceId = ucloud.String(resourceID)
	req.Port = ucloud.Int(d.Get("port").(int))
	resourceType := lbResourceTypeUHost
	if value, ok := d.GetOk("resource_type"); ok {
		resourceType = lbBackendCaseProdCvt.convert(value.(string))
	} else if len(strings.Split(resourceID, "-")) > 0 && strings.Split(resourceID, "-")[0] != eipResourceTypeUHost {
		return fmt.Errorf("must set `resource_type` when creating lb attachment")
	}
	req.ResourceType = ucloud.String(resourceType)

	resp, err := conn.AllocateBackend(req)
	if err != nil {
		return fmt.Errorf("error in create lb attachment, %s", err)
	}
	d.SetId(resp.BackendId)

	if _, err = lbAttachmentWaitForState(client, lbID, listenerID, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for lb attachment %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudLBAttachmentRead(d, meta)
}

func resourceUCloudLBAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating lb attachment, %s", err)
	}
	conn := client.ulbconn
	d.Partial(true)
	req := conn.NewUpdateBackendAttributeRequest()
	req.ULBId = ucloud.String(d.Get("load_balancer_id").(string))
	req.BackendId = ucloud.String(d.Id())
	if d.HasChange("port") && !d.IsNewResource() {
		req.Port = ucloud.Int(d.Get("port").(int))
		if _, err := conn.UpdateBackendAttribute(req); err != nil {
			return fmt.Errorf("error on %s to lb attachment %q, %s", "UpdateBackendAttribute", d.Id(), err)
		}
		d.SetPartial("port")
	}
	d.Partial(false)
	return resourceUCloudLBAttachmentRead(d, meta)
}

func resourceUCloudLBAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb attachment, %s", err)
	}
	var backendSet *ulb.ULBBackendSet
	lbID, lbOK := d.GetOk("load_balancer_id")
	listenerID, listenerOK := d.GetOk("listener_id")
	if lbOK && listenerOK {
		backendSet, err = client.describeBackendById(lbID.(string), listenerID.(string), d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb attachment %q, %s", d.Id(), err)
		}
		_ = d.Set("load_balancer_id", lbID)
		_ = d.Set("listener_id", listenerID)
	} else {
		backendSet, lbID, listenerID, err = client.describeBackendByOneId(d.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb attachment %q, %s", d.Id(), err)
		}
		_ = d.Set("load_balancer_id", lbID)
		_ = d.Set("listener_id", listenerID)
	}

	_ = d.Set("resource_id", backendSet.ResourceId)
	_ = d.Set("port", backendSet.Port)
	_ = d.Set("private_ip", backendSet.PrivateIP)
	_ = d.Set("status", lbAttachmentStatusCvt.convert(backendSet.Status))
	if value, ok := d.GetOk("resource_type"); ok && isStringIn(value.(string), []string{resourceTypeInstance}) {
		_ = d.Set("resource_type", lbBackendCaseProdCvt.unconvert(backendSet.ResourceType))
	} else {
		_ = d.Set("resource_type", backendSet.ResourceType)
	}
	return nil
}

func resourceUCloudLBAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb attachment, %s", err)
	}
	conn := client.ulbconn
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	req := conn.NewReleaseBackendRequest()
	req.ULBId = ucloud.String(lbID)
	req.BackendId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.describeBackendById(lbID, listenerID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb attachment before deleting %q, %s", d.Id(), err))
		}
		if _, err := conn.ReleaseBackend(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb attachment %q, %s", d.Id(), err))
		}
		if _, err := client.describeBackendById(lbID, listenerID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb attachment when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified lb attachment %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbAttachmentWaitForState(client *productClient, lbID, listenerID, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			backendSet, err := client.describeBackendById(lbID, listenerID, id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return backendSet, statusInitialized, nil
		},
	}
}
