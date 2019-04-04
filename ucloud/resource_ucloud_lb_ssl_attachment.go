package ucloud

import (
	"fmt"
	"strings"
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

	p := strings.Split(d.Id(), ":")

	sslAtSet, err := client.describeLBSSLAttachmentById(p[0], p[1], p[2])
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

	p := strings.Split(d.Id(), ":")

	req := conn.NewUnbindSSLRequest()
	req.SSLId = ucloud.String(p[0])
	req.ULBId = ucloud.String(p[1])
	req.VServerId = ucloud.String(p[2])

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.UnbindSSL(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb ssl attachment %q, %s", d.Id(), err))
		}

		_, err := client.describeLBSSLAttachmentById(p[0], p[1], p[2])
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb ssl attachment when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb ssl attachment %q has not been deleted due to unknown error", d.Id()))
	})
}
