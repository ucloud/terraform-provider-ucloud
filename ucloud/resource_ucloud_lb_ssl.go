package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBSSL() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBSSLCreate,
		Read:   resourceUCloudLBSSLRead,
		Delete: resourceUCloudLBSSLDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resource.PrefixedUniqueId("tf-lb-ssl"),
				ForceNew:     true,
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
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	req := conn.NewCreateSSLRequest()
	req.SSLType = ucloud.String("Pem")
	req.SSLName = ucloud.String(d.Get("name").(string))
	req.PrivateKey = ucloud.String(d.Get("private_key").(string))
	req.UserCert = ucloud.String(d.Get("user_cert").(string))

	if val, ok := d.GetOk("ca_cert"); ok {
		req.CaCert = ucloud.String(val.(string))
	}

	resp, err := conn.CreateSSL(req)
	if err != nil {
		return fmt.Errorf("error on creating lb SSL, %s", err)
	}

	d.SetId(resp.SSLId)

	return resourceUCloudLBSSLRead(d, meta)
}
func resourceUCloudLBSSLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	sslSet, err := client.describeLBSSLById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb %s, %s", d.Id(), err)
	}

	d.Set("name", sslSet.SSLName)
	d.Set("create_time", timestampToString(sslSet.CreateTime))

	return nil
}

func resourceUCloudLBSSLDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	req := conn.NewDeleteSSLRequest()
	req.SSLId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteSSL(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb ssl %s, %s", d.Id(), err))
		}

		_, err := client.describeLBSSLById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb ssl when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb ssl %s has not been deleted due to unknown error", d.Id()))
	})
}
