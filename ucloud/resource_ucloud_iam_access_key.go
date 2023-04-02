package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudIAMAccessKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMAccessKeyCreate,
		Update: resourceUCloudIAMAccessKeyUpdate,
		Read:   resourceUCloudIAMAccessKeyRead,
		Delete: resourceUCloudIAMAccessKeyDelete,

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret_file": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iamStatusActive,
				ValidateFunc: validation.StringInSlice([]string{iamStatusActive, iamStatusInactive}, false),
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudIAMAccessKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewCreateAccessKeyRequest()
	req.UserName = ucloud.String(d.Get("user_name").(string))

	resp, err := conn.CreateAccessKey(req)
	if err != nil {
		return fmt.Errorf("error on creating access key, %s", err)
	}

	if v, ok := d.GetOk("pgp_key"); ok {
		pgpKey := v.(string)
		encryptionKey, err := encryption.RetrieveGPGKey(pgpKey)
		if err != nil {
			return fmt.Errorf("error on retrieve gpg key, %s", err)
		}
		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, resp.AccessKey.AccessKeySecret, "UCloud IAM Access Key Secret")
		if err != nil {
			return fmt.Errorf("error on Encrypt Value, %s", err)
		}
		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret", encrypted)
	} else {
		if err := d.Set("secret", resp.AccessKey.AccessKeySecret); err != nil {
			return fmt.Errorf("fail to set secret, %s", err)
		}
	}
	if output, ok := d.GetOk("secret_file"); ok && output != nil {
		// create a secret_file and write access key to it.
		writeToFile(output.(string), resp.AccessKey)
	}

	d.SetId(resp.AccessKey.AccessKeyID)
	return resourceUCloudIAMAccessKeyUpdate(d, meta)
}

func resourceUCloudIAMAccessKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewUpdateAccessKeyRequest()
	req.AccessKeyID = ucloud.String(d.Id())
	req.Status = ucloud.String(d.Get("status").(string))

	if d.HasChange("status") {
		_, err := conn.UpdateAccessKey(req)
		if err != nil {
			return fmt.Errorf("error on update access key, %s", err)
		}
	}
	return resourceUCloudIAMAccessKeyRead(d, meta)
}

func resourceUCloudIAMAccessKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	accessKey, err := client.describeAccessKey(d.Get("user_name").(string), d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading access key %q, %s", d.Id(), err)
	}
	d.Set("user_name", d.Get("user_name").(string))
	d.Set("status", accessKey.Status)

	return nil
}

func resourceUCloudIAMAccessKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.iamconn

	req := conn.NewDeleteAccessKeyRequest()
	req.AccessKeyID = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if d.Get("status").(string) == iamStatusInactive {
			updateReq := conn.NewUpdateAccessKeyRequest()
			updateReq.AccessKeyID = ucloud.String(d.Id())
			updateReq.Status = ucloud.String(d.Get("status").(string))
			_, err := conn.UpdateAccessKey(updateReq)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error on update access key %q, %s", d.Id(), err))
			}
		}

		if _, err := conn.DeleteAccessKey(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting access key %q, %s", d.Id(), err))
		}

		_, err := client.describeAccessKey(d.Get("user_name").(string), d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading access key when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified access key %q has not been deleted due to unknown error", d.Id()))
	})

}
