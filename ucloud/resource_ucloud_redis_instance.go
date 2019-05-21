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

func resourceUCloudRedisInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudRedisInstanceCreate,
		Read:   resourceUCloudRedisInstanceRead,
		Update: resourceUCloudRedisInstanceUpdate,
		Delete: resourceUCloudRedisInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(
			diffValidateRedisInstanceTypeAndEngineVersion,
			customdiff.ValidateChange("instance_type", diffValidateRedisInstanceType),
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
				ValidateFunc: validateRedisInstanceType,
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"3.0",
					"3.2",
					"4.0",
				}, false),
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

			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validateKVStoreInstancePassword,
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

func resourceUCloudRedisInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	// skip error, because it has been validated at schema
	t, _ := parseRedisInstanceType(d.Get("instance_type").(string))

	if t.Type == "master" {
		return createActiveStandbyRedisInstance(d, meta)
	}
	return createDistributedRedisInstance(d, meta)
	// return fmt.Errorf("error on creating redis instance, unexpected type of instance type")
}

func resourceUCloudRedisInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	// skip error, because it has been validated at schema
	t, _ := parseRedisInstanceType(d.Get("instance_type").(string))

	if t.Type == "master" {
		return updateActiveStandbyRedisInstance(d, meta)
	}

	return updateDistributedRedisInstance(d, meta)
}

func resourceUCloudRedisInstanceRead(d *schema.ResourceData, meta interface{}) error {
	t, _ := parseRedisInstanceType(d.Get("instance_type").(string))

	if t.Type == "master" {
		return readActiveStandbyRedisInstance(d, meta)
	}

	return readDistributedRedisInstance(d, meta)
}

func resourceUCloudRedisInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	// skip error, because it has been validated at schema
	t, _ := parseRedisInstanceType(d.Get("instance_type").(string))

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if t.Type == "master" {
			return deleteActiveStandbyRedisInstance(d, meta)
		}

		return deleteDistributedRedisInstance(d, meta)
	})
}

func createActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	req := conn.NewCreateURedisGroupRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(getRedisCapability(d.Get("instance_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.HighAvailability = ucloud.String("enable")

	if v, ok := d.GetOk("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-redis-instance-"))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		req.Version = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("password"); ok {
		req.Password = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	// set default value of parametergroup
	parameterGroupId, err := setRedisDefaultParameterGroup(d, client)
	if err != nil {
		return err
	} else {
		req.ConfigId = ucloud.String(parameterGroupId)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(v.(string))
	}

	resp, err := conn.CreateURedisGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating redis instance, %s", err)
	}

	d.SetId(resp.GroupId)

	if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for redis instance %q complete creating, %s", d.Id(), err)
	}

	return readActiveStandbyRedisInstance(d, meta)
}

func createDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	req := conn.NewCreateUMemSpaceRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(getRedisCapability(d.Get("instance_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Protocol = ucloud.String("redis")

	if v, ok := d.GetOk("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-redis-instance-"))
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

	resp, err := conn.CreateUMemSpace(req)
	if err != nil {
		return fmt.Errorf("error on creating redis instance, %s", err)
	}

	d.SetId(resp.SpaceId)

	if err := client.waitDistributedRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for redis instance %q complete creating, %s", d.Id(), err)
	}

	return readDistributedRedisInstance(d, meta)
}

func updateActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) error {
	if err := updateActiveStandbyRedisInstanceWithoutRead(d, meta); err != nil {
		return err
	}
	return readActiveStandbyRedisInstance(d, meta)
}

func updateActiveStandbyRedisInstanceWithoutRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyURedisGroupNameRequest()
		req.GroupId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyURedisGroupName(req)
		if err != nil {
			return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyURedisGroupName", d.Id(), err)
		}
		d.SetPartial("name")

		if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyURedisGroupName", d.Id(), err)
		}
	}

	if d.HasChange("instance_type") && !d.IsNewResource() {
		req := conn.NewResizeURedisGroupRequest()
		req.GroupId = ucloud.String(d.Id())
		req.Size = ucloud.Int(getRedisCapability(d.Get("instance_type").(string)))

		_, err := conn.ResizeURedisGroup(req)
		if err != nil {
			return fmt.Errorf("error on %s to redis instance %q, %s", "ResizeURedisGroup", d.Id(), err)
		}
		d.SetPartial("instance_type")

		if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ResizeURedisGroup", d.Id(), err)
		}
	}

	if d.HasChange("password") && !d.IsNewResource() {
		password := d.Get("password").(string)

		req := client.pumemconn.NewModifyURedisGroupPasswordRequest()
		req.GroupId = ucloud.String(d.Id())
		req.Password = ucloud.String(password)

		_, err := client.pumemconn.ModifyURedisGroupPassword(req)
		if err != nil {
			return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyURedisGroupPassword", d.Id(), err)
		}
		d.SetPartial("password")

		if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyURedisGroupPassword", d.Id(), err)
		}
	}

	d.Partial(false)
	return nil
}

func updateDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyUMemSpaceNameRequest()
		req.SpaceId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyUMemSpaceName(req)
		if err != nil {
			return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyUMemSpaceName", d.Id(), err)
		}
		d.SetPartial("name")

		if err := client.waitDistributedRedisRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyUMemSpaceName", d.Id(), err)
		}
	}

	if d.HasChange("instance_type") && !d.IsNewResource() {
		req := conn.NewResizeUMemSpaceRequest()
		req.SpaceId = ucloud.String(d.Id())
		req.Size = ucloud.Int(getRedisCapability(d.Get("instance_type").(string)))

		_, err := conn.ResizeUMemSpace(req)
		if err != nil {
			return fmt.Errorf("error on %s to redis instance %q, %s", "ResizeUMemSpace", d.Id(), err)
		}
		d.SetPartial("instance_type")

		if err := client.waitDistributedRedisRunning(d.Id()); err != nil {
			return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ResizeUMemSpace", d.Id(), err)
		}
	}

	d.Partial(false)

	return readDistributedRedisInstance(d, meta)
}

func readActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	inst, err := client.describeActiveStandbyRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading redis instance %q, %s", d.Id(), err)
	}

	d.Set("name", inst.Name)
	d.Set("tag", inst.Tag)
	d.Set("charge_type", upperCamelCvt.convert(inst.ChargeType))
	d.Set("instance_type", fmt.Sprintf("redis-master-%v", inst.Size))
	d.Set("vpc_id", inst.VPCId)
	d.Set("subnet_id", inst.SubnetId)

	d.Set("ip_set", []map[string]interface{}{{
		"ip":   inst.VirtualIP,
		"port": inst.Port,
	}})
	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	d.Set("status", inst.State)
	return nil
}

func readDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	inst, err := client.describeDistributedRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading redis instance %q, %s", d.Id(), err)
	}

	d.Set("name", inst.Name)
	d.Set("tag", inst.Tag)
	d.Set("charge_type", upperCamelCvt.convert(inst.ChargeType))
	d.Set("instance_type", fmt.Sprintf("redis-distributed-%v", inst.Size))
	d.Set("vpc_id", inst.VPCId)
	d.Set("subnet_id", inst.SubnetId)

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

	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	d.Set("status", inst.State)
	return nil
}

func deleteActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) *resource.RetryError {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	req := conn.NewDeleteURedisGroupRequest()
	req.GroupId = ucloud.String(d.Id())

	if _, err := conn.DeleteURedisGroup(req); err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on deleting redis instance %q, %s", d.Id(), err))
	}

	_, err := client.describeActiveStandbyRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return resource.NonRetryableError(fmt.Errorf("error on reading redis instance when deleting %q, %s", d.Id(), err))
	}

	return resource.RetryableError(fmt.Errorf("the specified redis instance %q has not been deleted due to unknown error", d.Id()))
}

func deleteDistributedRedisInstance(d *schema.ResourceData, meta interface{}) *resource.RetryError {
	client := meta.(*UCloudClient)
	conn := client.umemconn

	req := conn.NewDeleteUMemSpaceRequest()
	req.SpaceId = ucloud.String(d.Id())
	if _, err := conn.DeleteUMemSpace(req); err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on deleting redis instance %q, %s", d.Id(), err))
	}

	_, err := client.describeDistributedRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return resource.NonRetryableError(fmt.Errorf("error on reading redis instance when deleting %q, %s", d.Id(), err))
	}
	return resource.RetryableError(fmt.Errorf("the specified redis instance %q has not been deleted due to unknown error", d.Id()))
}

func getRedisCapability(instType string) int {
	// skip error, because it has been validated at schema
	t, _ := parseRedisInstanceType(instType)
	return t.Memory
}

func setRedisDefaultParameterGroup(d *schema.ResourceData, client *UCloudClient) (string, error) {
	conn := client.pumemconn
	limit := 100
	offset := 0
	var parameterGroupId string
	for {
		req := conn.NewDescribeURedisConfigRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.Version = ucloud.String(d.Get("engine_version").(string))
		req.RegionFlag = ucloud.Bool(false)

		resp, err := conn.DescribeURedisConfig(req)
		if err != nil {
			return "", fmt.Errorf("error on reading redis parameter group when creating redis instance, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			return "", fmt.Errorf("error on querying defult value of redis parameter group")
		}

		for _, item := range resp.DataSet {
			if item.IsModify == "Unmodifiable" && item.State == "Usable" {
				parameterGroupId = item.ConfigId
				return parameterGroupId, nil
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}
	return "", fmt.Errorf("can not get the default redis parameter group")
}

func (c *UCloudClient) waitActiveStandbyRedisRunning(id string) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeActiveStandbyRedisById(id)
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

func (c *UCloudClient) waitDistributedRedisRunning(id string) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeDistributedRedisById(id)
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

func diffValidateRedisInstanceType(old, new, meta interface{}) error {
	if len(old.(string)) > 0 {
		oldType, _ := parseRedisInstanceType(old.(string))
		newType, _ := parseRedisInstanceType(new.(string))
		if newType.Type != oldType.Type {
			return fmt.Errorf("redis instance is not supported update the type of %q", "instance_type")
		}
		if newType.Engine != oldType.Engine {
			return fmt.Errorf("redis instance is not supported update the engine of %q", "instance_type")
		}
	}

	return nil
}

func diffValidateRedisInstanceTypeAndEngineVersion(diff *schema.ResourceDiff, v interface{}) error {
	engineVersion := diff.Get("engine_version").(string)
	redisType, _ := parseRedisInstanceType(diff.Get("instance_type").(string))

	if redisType.Type == "master" && engineVersion == "" {
		return fmt.Errorf("the %q argument must be set to active-standby redis instance", "engine_version")
	}

	if redisType.Type == "distributed" && engineVersion != "" {
		return fmt.Errorf("the %q argument is not apply to distributed redis instance", "engine_version")
	}

	return nil
}
