package ulb

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBSSL() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBSSLCreate,
		Read:   resourceUCloudLBSSLRead,
		Delete: resourceUCloudLBSSLDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},
			"private_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_cert": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ca_cert": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudLBSSLCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb ssl, %s", err)
	}
	conn := client.ulbconn
	req := conn.NewCreateSSLRequest()
	req.SSLType = ucloud.String("Pem")
	req.PrivateKey = ucloud.String(d.Get("private_key").(string))
	req.UserCert = ucloud.String(d.Get("user_cert").(string))
	if value, ok := d.GetOk("name"); ok {
		req.SSLName = ucloud.String(value.(string))
	} else {
		req.SSLName = ucloud.String(resource.PrefixedUniqueId("tf-lb-ssl-"))
	}
	if value, ok := d.GetOk("ca_cert"); ok {
		req.CaCert = ucloud.String(value.(string))
	}
	resp, err := conn.CreateSSL(req)
	if err != nil {
		return fmt.Errorf("error on creating lb SSL, %s", err)
	}
	d.SetId(resp.SSLId)
	return resourceUCloudLBSSLRead(d, meta)
}

func resourceUCloudLBSSLRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb ssl, %s", err)
	}
	sslSet, err := client.describeLBSSLById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb %q, %s", d.Id(), err)
	}
	_ = d.Set("name", sslSet.SSLName)
	_ = d.Set("create_time", timestampToString(sslSet.CreateTime))
	return nil
}

func resourceUCloudLBSSLDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb ssl, %s", err)
	}
	conn := client.ulbconn
	req := conn.NewDeleteSSLRequest()
	req.SSLId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteSSL(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb ssl %q, %s", d.Id(), err))
		}
		if _, err := client.describeLBSSLById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb ssl when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified lb ssl %q has not been deleted due to unknown error", d.Id()))
	})
}
