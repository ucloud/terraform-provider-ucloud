package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBAttachmentCreate,
		Read:   resourceUCloudLBAttachmentRead,
		Update: resourceUCloudLBAttachmentUpdate,
		Delete: resourceUCloudLBAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"listener_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"port": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      80,
				ValidateFunc: validateIntegerInRange(1, 65535),
			},

			"private_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudLBAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn

	req := conn.NewAllocateBackendRequest()
	req.ULBId = ucloud.String(d.Get("load_balancer_id").(string))
	req.VServerId = ucloud.String(d.Get("listener_id").(string))
	req.ResourceType = ucloud.String(uHostMap.convert(d.Get("resource_type").(string)))
	req.ResourceId = ucloud.String(d.Get("resource_id").(string))
	req.Port = ucloud.Int(d.Get("port").(int))

	resp, err := conn.AllocateBackend(req)
	if err != nil {
		return fmt.Errorf("error in create lb attachment, %s", err)
	}

	d.SetId(resp.BackendId)

	time.Sleep(10 * time.Second)

	return resourceUCloudLBAttachmentUpdate(d, meta)
}

func resourceUCloudLBAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	isChanged := false
	conn := meta.(*UCloudClient).ulbconn
	req := conn.NewUpdateBackendAttributeRequest()
	req.ULBId = ucloud.String(d.Get("load_balancer_id").(string))
	req.BackendId = ucloud.String(d.Id())

	if d.HasChange("port") && !d.IsNewResource() {
		isChanged = true
		req.Port = ucloud.Int(d.Get("port").(int))
		d.SetPartial("port")
	}

	if isChanged {
		_, err := conn.UpdateBackendAttribute(req)

		if err != nil {
			return fmt.Errorf("do %s failed in update lb attachment %s, %s", "UpdateBackendAttribute", d.Id(), err)
		}
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
		return fmt.Errorf("do %s failed in read lb attachment %s, %s", "DescribeVServer", d.Id(), err)
	}

	d.Set("resource_id", backendSet.ResourceId)
	d.Set("resource_type", uHostMap.unconvert(backendSet.ResourceType))
	d.Set("port", backendSet.Port)
	d.Set("private_ip", backendSet.PrivateIP)
	d.Set("status", attachmentStatus.transform(backendSet.Status))

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
			return resource.NonRetryableError(fmt.Errorf("error in delete lb attachment %s, %s", d.Id(), err))
		}

		_, err := client.describeBackendById(lbId, listenerId, d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete lb attachment %s, %s", "DescribeVServer", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete lb attachment but it still exists"))
	})
}
