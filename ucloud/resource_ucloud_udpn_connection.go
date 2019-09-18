package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/customdiff"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

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

func resourceUCloudUDPNConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udpnconn

	req := conn.NewAllocateUDPNRequest()
	req.Bandwidth = ucloud.Int(d.Get("bandwidth").(int))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	req.Peer1 = ucloud.String(client.region)
	req.Peer2 = ucloud.String(d.Get("peer_region").(string))

	resp, err := conn.AllocateUDPN(req)
	if err != nil {
		return fmt.Errorf("error on creating udpn connection, %s", err)
	}

	d.SetId(resp.UDPNId)

	// after create udpn connection, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    5 * time.Minute,
		Delay:      0 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			inst, err := client.describeDPNById(d.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return inst, statusInitialized, nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for udpn connection %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudUDPNConnectionRead(d, meta)
}

func resourceUCloudUDPNConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udpnconn

	d.Partial(true)

	if d.HasChange("bandwidth") && !d.IsNewResource() {
		req := conn.NewModifyUDPNBandwidthRequest()
		req.Region = ucloud.String(d.Get("peer_region").(string))
		req.UDPNId = ucloud.String(d.Id())
		req.Bandwidth = ucloud.Int(d.Get("bandwidth").(int))

		_, err := conn.ModifyUDPNBandwidth(req)
		if err != nil {
			return fmt.Errorf("error on %s to eip %q, %s", "ModifyUDPNBandwidth", d.Id(), err)
		}

		d.SetPartial("bandwidth")
	}

	d.Partial(false)

	return resourceUCloudUDPNConnectionRead(d, meta)
}

func resourceUCloudUDPNConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	inst, err := client.describeDPNById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading udpn connection %q, %s", d.Id(), err)
	}

	d.Set("bandwidth", inst.Bandwidth)
	d.Set("charge_type", upperCamelCvt.convert(inst.ChargeType))

	// peer1, peer2 has unordered from server response
	if inst.Peer1 == client.region {
		d.Set("peer_region", inst.Peer2)
	} else {
		d.Set("peer_region", inst.Peer1)
	}

	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	return nil
}

func resourceUCloudUDPNConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udpnconn

	req := conn.NewReleaseUDPNRequest()
	req.UDPNId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.describeDPNById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading udpn connection when deleting %q, %s", d.Id(), err))
		}

		_, err = conn.ReleaseUDPN(req)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting udpn connection %q, %s", d.Id(), err))
		}

		_, err = client.describeDPNById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading udpn connection when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified udpn connection %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateUDPNPeerRegion(old, new, meta interface{}) error {
	client := meta.(*UCloudClient)

	if new.(string) == client.region {
		return fmt.Errorf(
			"expected the peering region %q to be different with provider's region %q",
			new.(string), client.region,
		)
	}

	return nil
}
