package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

const (
	lbResourceTypeUHost = "UHost"
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				Deprecated:   "attribute `resource_type` is deprecated for optimizing parameters",
				ValidateFunc: validation.StringInSlice([]string{"instance"}, false),
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
	req.ResourceType = ucloud.String(lbResourceTypeUHost)
	req.ResourceId = ucloud.String(d.Get("resource_id").(string))
	req.Port = ucloud.Int(d.Get("port").(int))

	resp, err := conn.AllocateBackend(req)
	if err != nil {
		return fmt.Errorf("error in create lb attachment, %s", err)
	}

	d.SetId(resp.BackendId)

	// after create lb attachment, we need to wait it initialized
	stateConf := lbAttachmentWaitForState(client, lbId, listenerId, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for lb attachment %s complete creating, %s", d.Id(), err)
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
			return fmt.Errorf("error on %s to lb attachment %s, %s", "UpdateBackendAttribute", d.Id(), err)
		}

		d.SetPartial("port")
	}

	d.Partial(false)

	return resourceUCloudLBAttachmentRead(d, meta)
}

func resourceUCloudLBAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)

	backendSet, err := client.describeBackendById(lbId, listenerId, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb attachment %s, %s", d.Id(), err)
	}

	d.Set("resource_id", backendSet.ResourceId)
	d.Set("resource_type", titleCaseProdCvt.unconvert(backendSet.ResourceType))
	d.Set("port", backendSet.Port)
	d.Set("private_ip", backendSet.PrivateIP)
	d.Set("status", lbAttachmentStatusCvt.convert(backendSet.Status))

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
		if _, err := conn.ReleaseBackend(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb attachment %s, %s", d.Id(), err))
		}

		_, err := client.describeBackendById(lbId, listenerId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb attachment when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb attachment %s has not been deleted due to unknown error", d.Id()))
	})
}

func lbAttachmentWaitForState(client *UCloudClient, lbId, listenerId, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {
			backendSet, err := client.describeBackendById(lbId, listenerId, id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			state := lbAttachmentStatusCvt.convert(backendSet.Status)
			if state != "normalRunning" {
				state = statusPending
			} else {
				state = statusInitialized
			}

			return backendSet, state, nil
		},
	}
}
