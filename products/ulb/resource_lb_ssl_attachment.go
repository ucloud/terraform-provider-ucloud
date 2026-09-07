package ulb

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb ssl attachment, %s", err)
	}
	conn := client.ulbconn
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	sslID := d.Get("ssl_id").(string)
	req := conn.NewBindSSLRequest()
	req.ULBId = ucloud.String(lbID)
	req.VServerId = ucloud.String(listenerID)
	req.SSLId = ucloud.String(sslID)
	if _, err := conn.BindSSL(req); err != nil {
		return fmt.Errorf("error in create lb ssl attachment, %s", err)
	}
	d.SetId(fmt.Sprintf("%s:%s:%s", sslID, lbID, listenerID))
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusInitialized},
		Refresh: func() (interface{}, string, error) {
			sslSet, err := client.describeLBSSLAttachmentById(sslID, lbID, listenerID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return sslSet, statusInitialized, nil
		},
		Timeout:    2 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for ssl attachment %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudLBSSLAttachmentRead(d, meta)
}

func resourceUCloudLBSSLAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb ssl attachment, %s", err)
	}
	p := strings.Split(d.Id(), ":")
	sslAtSet, err := client.describeLBSSLAttachmentById(p[0], p[1], p[2])
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb ssl attachment %q, %s", d.Id(), err)
	}
	_ = d.Set("load_balancer_id", sslAtSet.ULBId)
	_ = d.Set("listener_id", sslAtSet.VServerId)
	_ = d.Set("ssl_id", p[0])
	return nil
}

func resourceUCloudLBSSLAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb ssl attachment, %s", err)
	}
	conn := client.ulbconn
	p := strings.Split(d.Id(), ":")
	if _, err := client.describeLBSSLAttachmentById(p[0], p[1], p[2]); err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb ssl attachment before deleting %q, %s", d.Id(), err)
	}
	req := conn.NewUnbindSSLRequest()
	req.SSLId = ucloud.String(p[0])
	req.ULBId = ucloud.String(p[1])
	req.VServerId = ucloud.String(p[2])
	if _, err := conn.UnbindSSL(req); err != nil {
		return fmt.Errorf("error on deleting lb ssl attachment %q, %s", d.Id(), err)
	}
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusInitialized},
		Refresh: func() (interface{}, string, error) {
			sslSet, err := client.describeLBSSLAttachmentById(p[0], p[1], p[2])
			if err != nil {
				if isNotFoundError(err) {
					return sslSet, statusInitialized, nil
				}
				return nil, "", err
			}
			return sslSet, statusPending, nil
		},
		Timeout:    2 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for ssl attachment %q complete deleting, %s", d.Id(), err)
	}
	return nil
}
