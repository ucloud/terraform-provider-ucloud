package udpn

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUDPNConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUDPNConnectionCreate,
		Read:   resourceUCloudUDPNConnectionRead,
		Update: resourceUCloudUDPNConnectionUpdate,
		Delete: resourceUCloudUDPNConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("peer_region", diffValidateUDPNPeerRegion),
		),

		Schema: map[string]*schema.Schema{
			"bandwidth": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(2, 1000),
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "month",
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"dynamic",
				}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"peer_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudUDPNConnectionCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating udpn connection, %s", err)
	}
	conn := client

	req := conn.NewAllocateUDPNRequest()
	req.Bandwidth = ucloud.Int(data.Get("bandwidth").(int))
	req.ChargeType = ucloud.String(upperCamelUnconvert(data.Get("charge_type").(string)))

	if value, ok := data.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(value.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	req.Peer1 = ucloud.String(providerRegion(client))
	req.Peer2 = ucloud.String(data.Get("peer_region").(string))

	resp, err := conn.AllocateUDPN(req)
	if err != nil {
		return fmt.Errorf("error on creating udpn connection, %s", err)
	}

	data.SetId(resp.UDPNId)

	// after create udpn connection, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    5 * time.Minute,
		Delay:      0 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			instance, err := describeDPNById(client, data.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return instance, statusInitialized, nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for udpn connection %q complete creating, %s", data.Id(), err)
	}

	return resourceUCloudUDPNConnectionRead(data, meta)
}

func resourceUCloudUDPNConnectionUpdate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating udpn connection, %s", err)
	}
	conn := client

	data.Partial(true)

	if data.HasChange("bandwidth") && !data.IsNewResource() {
		req := conn.NewModifyUDPNBandwidthRequest()
		req.Region = ucloud.String(data.Get("peer_region").(string))
		req.UDPNId = ucloud.String(data.Id())
		req.Bandwidth = ucloud.Int(data.Get("bandwidth").(int))

		_, err := conn.ModifyUDPNBandwidth(req)
		if err != nil {
			return fmt.Errorf("error on %s to eip %q, %s", "ModifyUDPNBandwidth", data.Id(), err)
		}

		data.SetPartial("bandwidth")
	}

	data.Partial(false)

	return resourceUCloudUDPNConnectionRead(data, meta)
}

func resourceUCloudUDPNConnectionRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading udpn connection, %s", err)
	}

	instance, err := describeDPNById(client, data.Id())
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading udpn connection %q, %s", data.Id(), err)
	}

	_ = data.Set("bandwidth", instance.Bandwidth)
	_ = data.Set("charge_type", upperCamelConvert(instance.ChargeType))

	// peer1, peer2 has unordered from server response
	if instance.Peer1 == providerRegion(client) {
		_ = data.Set("peer_region", instance.Peer2)
	} else {
		_ = data.Set("peer_region", instance.Peer1)
	}

	_ = data.Set("create_time", timestampToString(instance.CreateTime))
	_ = data.Set("expire_time", timestampToString(instance.ExpireTime))
	return nil
}

func resourceUCloudUDPNConnectionDelete(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting udpn connection, %s", err)
	}
	conn := client

	req := conn.NewReleaseUDPNRequest()
	req.UDPNId = ucloud.String(data.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := describeDPNById(client, data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading udpn connection when deleting %q, %s", data.Id(), err))
		}

		_, err = conn.ReleaseUDPN(req)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting udpn connection %q, %s", data.Id(), err))
		}

		_, err = describeDPNById(client, data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading udpn connection when deleting %q, %s", data.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified udpn connection %q has not been deleted due to unknown error", data.Id()))
	})
}

func diffValidateUDPNPeerRegion(old, new, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return err
	}

	if new.(string) == providerRegion(client) {
		return fmt.Errorf(
			"expected the peering region %q to be different with provider's region %q",
			new.(string), providerRegion(client),
		)
	}

	return nil
}
