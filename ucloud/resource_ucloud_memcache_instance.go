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

func resourceUCloudMemcacheInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudMemcacheInstanceCreate,
		Read:   resourceUCloudMemcacheInstanceRead,
		Update: resourceUCloudMemcacheInstanceUpdate,
		Delete: resourceUCloudMemcacheInstanceDelete,

		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("instance_type", diffValidateMemcacheInstanceType),
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateKVStoreInstanceName,
			},

			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateMemcacheInstanceType,
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

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"ip_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
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

func resourceUCloudMemcacheInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	req := conn.NewCreateUMemcacheGroupRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(getMemcacheCapability(d.Get("instance_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Protocol = ucloud.String("memcache")

	if v, ok := d.GetOk("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-memcache-instance-"))
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	resp, err := conn.CreateUMemcacheGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating memcache instance, %s", err)
	}

	d.SetId(resp.GroupId)

	if err := client.waitActiveStandbyMemcacheRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for memcache instance %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudMemcacheInstanceRead(d, meta)
}

func resourceUCloudMemcacheInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.pumemconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyUMemcacheGroupNameRequest()
		req.GroupId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyUMemcacheGroupName(req)
		if err != nil {
			return fmt.Errorf("error on %s to memcache instance %q, %s", "ModifyUMemcacheGroupName", d.Id(), err)
		}

		if err := client.waitActiveStandbyMemcacheRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to memcache instance %q, %s", "ModifyUMemcacheGroupName", d.Id(), err)
		}
	}

	if d.HasChange("instance_type") && !d.IsNewResource() {
		req := conn.NewResizeUMemcacheGroupRequest()
		req.GroupId = ucloud.String(d.Id())
		req.Size = ucloud.Int(getMemcacheCapability(d.Get("instance_type").(string)))

		_, err := conn.ResizeUMemcacheGroup(req)
		if err != nil {
			return fmt.Errorf("error on %s to memcache instance %q, %s", "ResizeUMemcacheGroup", d.Id(), err)
		}

		if err := client.waitActiveStandbyMemcacheRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to memcache instance %q, %s", "ResizeUMemcacheGroup", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudMemcacheInstanceRead(d, meta)
}

func resourceUCloudMemcacheInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	inst, err := client.describeActiveStandbyMemcacheById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading memcache instance %q, %s", d.Id(), err)
	}

	d.Set("name", inst.Name)
	d.Set("tag", inst.Tag)
	d.Set("charge_type", upperCamelCvt.convert(inst.ChargeType))
	d.Set("instance_type", fmt.Sprintf("memcache-master-%v", inst.Size))
	d.Set("vpc_id", inst.VPCId)
	d.Set("subnet_id", inst.SubnetId)

	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	d.Set("status", inst.State)

	ipSet := []map[string]interface{}{}
	for _, addr := range inst.Address {
		ipItem := map[string]interface{}{
			"ip":   addr.IP,
			"port": addr.Port,
		}
		ipSet = append(ipSet, ipItem)
	}

	if err := d.Set("ip_set", ipSet); err != nil {
		return err
	}
	return nil
}

func resourceUCloudMemcacheInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn
	req := conn.NewDeleteUMemcacheGroupRequest()
	req.GroupId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteUMemcacheGroup(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting memcache instance %q, %s", d.Id(), err))
		}

		_, err := client.describeActiveStandbyMemcacheById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading memcache instance when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified memcache instance %q has not been deleted due to unknown error", d.Id()))
	})
}

func getMemcacheCapability(instType string) int {
	// skip error, because it has been validated at schema
	t, _ := parseMemcacheInstanceType(instType)
	return t.Memory
}

func diffValidateMemcacheInstanceType(old, new, meta interface{}) error {
	if len(old.(string)) > 0 {
		oldType, _ := parseMemcacheInstanceType(old.(string))
		newType, _ := parseMemcacheInstanceType(new.(string))
		if newType.Type != oldType.Type {
			return fmt.Errorf("memcache instance is not supported update the type of %q", "instance_type")
		}
		if newType.Engine != oldType.Engine {
			return fmt.Errorf("memcache instance is not supported update the engine of %q", "instance_type")
		}
	}

	return nil
}

func (c *UCloudClient) waitActiveStandbyMemcacheRunning(id string) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeActiveStandbyMemcacheById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		if resp.State != upperCamelCvt.unconvert(statusRunning) {
			return nil, statusPending, nil
		}
		return resp, "ok", nil
	}

	return waitForMemoryInstance(refresh)
}
