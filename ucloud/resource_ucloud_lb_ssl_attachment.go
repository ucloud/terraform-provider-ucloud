package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBSSLAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBSSLAttachmentCreate,
		Read:   resourceUCloudLBSSLAttachmentRead,
		Delete: resourceUCloudLBSSLAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"ssl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

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
		},
	}
}

func resourceUCloudLBSSLAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	listenerId := d.Get("listener_id").(string)
	sslId := d.Get("ssl_id").(string)

	req := conn.NewBindSSLRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)
	req.SSLId = ucloud.String(sslId)

	if _, err := conn.BindSSL(req); err != nil {
		return fmt.Errorf("error in create lb ssl attachment, %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", sslId, lbId, listenerId))

	return resourceUCloudLBSSLAttachmentRead(d, meta)
}

func resourceUCloudLBSSLAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	attach, err := parseAttachmentInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing lb ssl attachment %q, %s", d.Id(), err)
	}

	sslAtSet, err := client.describeLBSSLAttachmentById(attach.PrimaryId, attach.SecondId, attach.ThirdId)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb ssl attachment %q, %s", d.Id(), err)
	}

	d.Set("load_balancer_id", sslAtSet.ULBId)
	d.Set("listener_id", sslAtSet.VServerId)

	return nil
}

func resourceUCloudLBSSLAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	attach, err := parseAttachmentInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing lb ssl attachment %q, %s", d.Id(), err)
	}

	req := conn.NewUnbindSSLRequest()
	req.SSLId = ucloud.String(attach.PrimaryId)
	req.ULBId = ucloud.String(attach.SecondId)
	req.VServerId = ucloud.String(attach.ThirdId)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.UnbindSSL(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb ssl attachment %q, %s", d.Id(), err))
		}

		_, err := client.describeLBSSLAttachmentById(attach.PrimaryId, attach.SecondId, attach.ThirdId)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb ssl attachment when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb ssl attachment %q has not been deleted due to unknown error", d.Id()))
	})
}
