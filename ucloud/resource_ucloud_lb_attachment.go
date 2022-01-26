package ucloud

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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if isStringIn(old, []string{resourceTypeInstance, lbResourceTypeUHost}) && isStringIn(new, []string{resourceTypeInstance, lbResourceTypeUHost}) {
						return true
					}
					return false
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
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	req := conn.NewAllocateBackendRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)
	resourceId := d.Get("resource_id").(string)
	req.ResourceId = ucloud.String(resourceId)
	req.Port = ucloud.Int(d.Get("port").(int))
	resourceType := lbResourceTypeUHost
	if v, ok := d.GetOk("resource_type"); ok {
		resourceType = lbBackendCaseProdCvt.convert(v.(string))
	} else if len(strings.Split(resourceId, "-")) > 0 && strings.Split(resourceId, "-")[0] != eipResourceTypeUHost {
		return fmt.Errorf("must set `resource_type` when creating lb attachment")
	}

	req.ResourceType = ucloud.String(resourceType)

	resp, err := conn.AllocateBackend(req)
	if err != nil {
		return fmt.Errorf("error in create lb attachment, %s", err)
	}

	d.SetId(resp.BackendId)

	// after create lb attachment, we need to wait it initialized
	stateConf := lbAttachmentWaitForState(client, lbId, listenerId, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for lb attachment %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudLBAttachmentRead(d, meta)
}

func resourceUCloudLBAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn

	d.Partial(true)

	req := conn.NewUpdateBackendAttributeRequest()
	req.ULBId = ucloud.String(d.Get("load_balancer_id").(string))
	req.BackendId = ucloud.String(d.Id())

	isChanged := false

	if d.HasChange("port") && !d.IsNewResource() {
		isChanged = true
		req.Port = ucloud.Int(d.Get("port").(int))
	}

	if isChanged {
		_, err := conn.UpdateBackendAttribute(req)
		if err != nil {
			return fmt.Errorf("error on %s to lb attachment %q, %s", "UpdateBackendAttribute", d.Id(), err)
		}

		d.SetPartial("port")
	}

	d.Partial(false)

	return resourceUCloudLBAttachmentRead(d, meta)
}

func resourceUCloudLBAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	var err error
	var backendSet *ulb.ULBBackendSet
	lbId, lbOk := d.GetOk("load_balancer_id")
	listenerId, lsOk := d.GetOk("listener_id")

	if lbOk && lsOk {
		backendSet, err = client.describeBackendById(lbId.(string), listenerId.(string), d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb attachment %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", lbId)
		d.Set("listener_id", listenerId)
	} else {
		backendSet, lbId, listenerId, err = client.describeBackendByOneId(d.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb attachment %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", lbId)
		d.Set("listener_id", listenerId)
	}

	d.Set("resource_id", backendSet.ResourceId)
	d.Set("port", backendSet.Port)
	d.Set("private_ip", backendSet.PrivateIP)
	d.Set("status", lbAttachmentStatusCvt.convert(backendSet.Status))

	if v, ok := d.GetOk("resource_type"); ok && isStringIn(v.(string), []string{resourceTypeInstance}) {
		d.Set("resource_type", lbBackendCaseProdCvt.unconvert(backendSet.ResourceType))
	} else {
		d.Set("resource_type", backendSet.ResourceType)
	}

	return nil
}

func resourceUCloudLBAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	req := conn.NewReleaseBackendRequest()
	req.ULBId = ucloud.String(lbId)
	req.BackendId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.describeBackendById(lbId, listenerId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb attachment before deleting %q, %s", d.Id(), err))
		}

		if _, err := conn.ReleaseBackend(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb attachment %q, %s", d.Id(), err))
		}

		_, err = client.describeBackendById(lbId, listenerId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb attachment when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb attachment %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbAttachmentWaitForState(client *UCloudClient, lbId, listenerId, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			backendSet, err := client.describeBackendById(lbId, listenerId, id)
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
