package umem

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudRedisInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudRedisInstanceCreate,
		Read:   resourceUCloudRedisInstanceRead,
		Update: resourceUCloudRedisInstanceUpdate,
		Delete: resourceUCloudRedisInstanceDelete,

		CustomizeDiff: customdiff.All(
			diffValidateRedisInstanceTypeAndEngineVersion,
			diffValidateRedisStandbyZone,
			diffValidateBackup,
			diffValidateBlockCnt,
			diffValidateRedisRestart,
			customdiff.ValidateChange("instance_type", diffValidateRedisInstanceType),
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"standby_zone": {
				Type:     schema.TypeString,
				Optional: true,
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

			// block_cnt is the number of shards (blocks) for a distributed redis
			// instance, sent as CreateUMemSpace's BlockCnt. The backend now requires
			// it, so when omitted we send defaultDistributedRedisBlockCnt. It is
			// create-only and ForceNew: shard count cannot be changed in place (a
			// capacity resize via ResizeUMemSpace keeps the existing shards). Only
			// valid for the distributed instance type; rejected for active-standby.
			// The live shards are exposed via the computed block_set attribute.
			"block_cnt": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
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
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"auto_backup": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"enable",
					"disable",
				}, false),
				Computed: true,
			},

			"backup_begin_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},

			// restart_trigger is an imperative, one-shot knob: changing its value
			// restarts an active-standby redis. It is not read back from the API.
			"restart_trigger": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// transform_type is a optionalknob:changing redis instance state from "Running" or "ISolation"
			// if user wants to shutdown a redis active-standby instance, sets transform_type to "UNBind"
			// if user wants to start a redis active-standby instance, sets transform_type to "Bind"
			"transform_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"UNBind", "Bind"}, false),
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

			// block_set exposes the shard (block) information of a distributed redis
			// instance. It is only populated for the distributed instance type and
			// stays empty for active-standby redis.
			"block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"block_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"block_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_read_weight": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_slot_begin": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_slot_end": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_state": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"block_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"block_used_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"block_vip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			// read_mode is the cluster level read/write splitting strategy of a
			// distributed redis instance, reported alongside the shard information.
			"read_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// proxy_set exposes the proxy information of a distributed redis instance.
			// It is only populated for the distributed instance type and stays empty
			// for active-standby redis.
			"proxy_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"proxy_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"vip": {
							Type:     schema.TypeString,
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

	if v, ok := d.GetOk("standby_zone"); ok {
		return fmt.Errorf("standby_zone %q only be supported for Active-Standby Redis, not be supported for Distributed Redis", v)
	}
	return createDistributedRedisInstance(d, meta)
}

func resourceUCloudRedisInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	// skip error, because it has been validated at schema
	t, _ := parseRedisInstanceType(d.Get("instance_type").(string))

	if t.Type == "master" {
		return updateActiveStandbyRedisInstance(d, meta)
	}

	if v, ok := d.GetOk("standby_zone"); ok {
		return fmt.Errorf("standby_zone %q only be supported for Active-Standby Redis, not be supported for Distributed Redis", v)
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating redis instance, %s", err)
	}
	conn := client.umemconn

	req := conn.NewCreateURedisGroupRequest()
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Size = ucloud.Int(getRedisCapability(d.Get("instance_type").(string)))
	req.HighAvailability = ucloud.String("enable")
	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if val, ok := d.GetOk("standby_zone"); ok {
		req.SlaveZone = ucloud.String(val.(string))
	}

	if v, ok := d.GetOkExists("duration"); ok {
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
	parameterGroupId, err := getRedisDefaultParameterGroup(d, client)
	if err != nil {
		return err
	} else {
		req.ConfigId = ucloud.String(parameterGroupId)
	}

	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	if v, ok := d.GetOkExists("backup_begin_time"); ok {
		req.BackupTime = ucloud.Int(v.(int))
	} else {
		req.BackupTime = ucloud.Int(3)
	}

	if v, ok := d.GetOk("auto_backup"); ok {
		req.AutoBackup = ucloud.String(v.(string))
	} else {
		req.AutoBackup = ucloud.String("disable")
	}

	resp, err := conn.CreateURedisGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating redis instance, %s", err)
	}

	d.SetId(resp.GroupId)

	if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for redis instance %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudRedisInstanceUpdate(d, meta)
}

func createDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating redis instance, %s", err)
	}
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

	if v, ok := d.GetOk("password"); ok {
		req.Password = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-redis-instance-"))
	}

	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))

	// BlockCnt (shard count) is required by the backend for distributed redis.
	// Use the configured block_cnt, falling back to the documented default.
	if v, ok := d.GetOk("block_cnt"); ok {
		req.BlockCnt = ucloud.Int(v.(int))
	} else {
		req.BlockCnt = ucloud.Int(defaultDistributedRedisBlockCnt)
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

	return resourceUCloudRedisInstanceUpdate(d, meta)
}

func updateActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) error {
	if err := updateActiveStandbyRedisInstanceWithoutRead(d, meta); err != nil {
		return err
	}
	return readActiveStandbyRedisInstance(d, meta)
}

func updateActiveStandbyRedisInstanceWithoutRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating redis instance, %s", err)
	}

	d.Partial(true)

	// Each branch applies one in-place change through its own helper following
	// the same call -> SetPartial -> wait shape, and blocks until the instance
	// settles again. Only transform/restart below are order-sensitive.

	if d.HasChange("name") && !d.IsNewResource() {
		if err := renameActiveStandbyRedis(client, d); err != nil {
			return err
		}
	}

	if d.HasChange("instance_type") && !d.IsNewResource() {
		if err := resizeActiveStandbyRedis(client, d); err != nil {
			return err
		}
	}

	if d.HasChange("password") && !d.IsNewResource() {
		if err := modifyActiveStandbyRedisPassword(client, d); err != nil {
			return err
		}
	}

	// UpdateURedisBackupStrategy takes the full strategy, so a change to either
	// attribute re-sends both
	if (d.HasChange("backup_begin_time") || d.HasChange("auto_backup")) && !d.IsNewResource() {
		if err := updateActiveStandbyRedisBackup(client, d); err != nil {
			return err
		}
	}

	// Bind or UNBind must run BEFORE the restart below: RestartURedisGroup is
	// rejected by the backend (RetCode 21018) while the instance is isolated, so
	// a combined "Bind + restart" apply has to bring the instance back to Running
	// first.
	if d.HasChange("transform_type") && !d.IsNewResource() {
		if err := transformActiveStandbyRedis(client, d); err != nil {
			return err
		}
	}

	// A changed restart_trigger requests a one-shot restart. Skipped while the
	// resource is new so a freshly created instance is not restarted right away.
	if d.HasChange("restart_trigger") && !d.IsNewResource() {
		if err := restartActiveStandbyRedis(client, d); err != nil {
			return err
		}
	}

	d.Partial(false)
	return nil
}

// renameActiveStandbyRedis applies a name change through ModifyURedisGroupName.
func renameActiveStandbyRedis(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn

	req := conn.NewModifyURedisGroupNameRequest()
	req.GroupId = ucloud.String(d.Id())
	req.Name = ucloud.String(d.Get("name").(string))

	if _, err := conn.ModifyURedisGroupName(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyURedisGroupName", d.Id(), err)
	}
	d.SetPartial("name")

	if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyURedisGroupName", d.Id(), err)
	}
	return nil
}

// resizeActiveStandbyRedis applies an instance_type change through
// ResizeURedisGroup.
func resizeActiveStandbyRedis(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn
	targetSize := getRedisCapability(d.Get("instance_type").(string))

	req := conn.NewResizeURedisGroupRequest()
	req.GroupId = ucloud.String(d.Id())
	req.Size = ucloud.Int(targetSize)

	if _, err := conn.ResizeURedisGroup(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ResizeURedisGroup", d.Id(), err)
	}
	d.SetPartial("instance_type")

	// Wait for the target size, not just "Running": the resize is async and
	// right after ResizeURedisGroup returns the instance may still report the
	// old size with state Running for a moment before flipping to Resizing.
	// Waiting on Running alone returns early and the following read then
	// writes the stale size back into state.
	if err := client.waitActiveStandbyRedisResized(d.Id(), targetSize); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ResizeURedisGroup", d.Id(), err)
	}
	return nil
}

// modifyActiveStandbyRedisPassword applies a password change through
// ModifyURedisGroupPassword (pumemconn endpoint).
func modifyActiveStandbyRedisPassword(client *productClient, d *schema.ResourceData) error {
	req := client.pumemconn.NewModifyURedisGroupPasswordRequest()
	req.GroupId = ucloud.String(d.Id())
	req.Password = ucloud.String(d.Get("password").(string))

	if _, err := client.pumemconn.ModifyURedisGroupPassword(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyURedisGroupPassword", d.Id(), err)
	}
	d.SetPartial("password")

	if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyURedisGroupPassword", d.Id(), err)
	}
	return nil
}

// updateActiveStandbyRedisBackup applies an auto_backup / backup_begin_time
// change through UpdateURedisBackupStrategy. The API takes the full strategy,
// so both attributes are always sent; unset values fall back to the same
// defaults used at create time ("disable" / 3 o'clock).
func updateActiveStandbyRedisBackup(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn

	req := conn.NewUpdateURedisBackupStrategyRequest()
	req.GroupId = ucloud.String(d.Id())

	if v, ok := d.GetOkExists("backup_begin_time"); ok {
		req.BackupTime = ucloud.String(strconv.Itoa(v.(int)))
	} else {
		req.BackupTime = ucloud.String("3")
	}

	if v, ok := d.GetOk("auto_backup"); ok {
		req.AutoBackup = ucloud.String(v.(string))
	} else {
		req.AutoBackup = ucloud.String("disable")
	}

	if _, err := conn.UpdateURedisBackupStrategy(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "UpdateURedisBackupStrategy", d.Id(), err)
	}

	d.SetPartial("auto_backup")
	d.SetPartial("backup_begin_time")
	return nil
}

// transformActiveStandbyRedis applies a transform_type change through
// ISolationURedisGroup: "UNBind" shuts the instance down (target state
// ISolation), "Bind" starts it back up (target state Running). The API is
// async and the instance still reports the OLD state right after it returns,
// so the wait blocks until the target state of this direction is reached.
func transformActiveStandbyRedis(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn
	transformType := d.Get("transform_type").(string)

	req := conn.NewISolationURedisGroupRequest()
	req.GroupId = ucloud.String(d.Id())
	req.TransformType = ucloud.String(transformType)

	if _, err := conn.ISolationURedisGroup(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ISolationURedisGroup", d.Id(), err)
	}

	d.SetPartial("transform_type")

	if err := client.waitActiveStandbyRedisTransform(d.Id(), transformType); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ISolationURedisGroup", d.Id(), err)
	}
	return nil
}

// restartActiveStandbyRedis performs the one-shot restart requested through a
// changed restart_trigger. RestartURedisGroup only works on a Running instance;
// the backend rejects an isolated one with an opaque RetCode 21018, so the
// state is checked up front and a clear error is returned instead (any
// Bind/UNBind requested in the same apply has already been applied by the
// transform branch, which must run before this one).
func restartActiveStandbyRedis(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn

	inst, err := client.describeActiveStandbyRedisById(d.Id())
	if err != nil {
		return fmt.Errorf("error on reading redis instance %q before restart, %s", d.Id(), err)
	}
	if inst.State != statusRunning {
		return fmt.Errorf("cannot restart redis instance %q in state %q: the instance must be Running, set transform_type to %q to start an isolated instance first", d.Id(), inst.State, "Bind")
	}

	req := conn.NewRestartURedisGroupRequest()
	req.GroupId = ucloud.String(d.Id())

	if _, err := conn.RestartURedisGroup(req); err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "RestartURedisGroup", d.Id(), err)
	}

	// commit the new trigger before waiting, so a later failure will not re-fire the restart
	d.SetPartial("restart_trigger")

	// RestartURedisGroup is async; block until the instance is back to Running
	if err := client.waitActiveStandbyRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "RestartURedisGroup", d.Id(), err)
	}
	return nil
}

// ================================================Distributed Redis helpers========================================

func updateDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating redis instance, %s", err)
	}

	d.Partial(true)

	// Each branch applies one in-place change through its own helper following
	if d.HasChange("name") && !d.IsNewResource() {
		if err := renameDistributedRedis(client, d); err != nil {
			return err
		}
	}

	// The backend does not support changing the shard count of a distributed redis
	if d.HasChange("instance_type") && !d.IsNewResource() {
		if err := resizeDistributedRedisInstance(client, d); err != nil {
			return err
		}
		d.SetPartial("instance_type")
	}

	// password
	if d.HasChange("password") && !d.IsNewResource() {
		if err := modifyDistributedRedisPassword(client, d); err != nil {
			return err
		}
	}

	d.Partial(false)

	return readDistributedRedisInstance(d, meta)
}

// renameDistributedRedis applies a name change through ModifyUMemSpaceName.
func renameDistributedRedis(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn

	req := conn.NewModifyUMemSpaceNameRequest()
	req.SpaceId = ucloud.String(d.Id())
	req.Name = ucloud.String(d.Get("name").(string))
	// the UMemSpace API family validates the zone common param on the backend
	req.Zone = ucloud.String(d.Get("availability_zone").(string))

	_, err := conn.ModifyUMemSpaceName(req)
	if err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyUMemSpaceName", d.Id(), err)
	}
	d.SetPartial("name")

	if err := client.waitDistributedRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyUMemSpaceName", d.Id(), err)
	}
	return nil
}

// modifyDistributedRedisPassword applies a password change through ModifyUMemPassword.
// Two API-family quirks must be handled here:
//   - Like the rest of the UMemSpace family, ModifyUMemPassword validates the
//     zone common param on the backend ("Missing Params [zone_id]" without it),
//     so Zone is always sent even though the schema marks only SpaceId/Password
//     as required.
//   - Unlike CreateUMemSpace (and unlike the private SDK's
//     ModifyURedisGroupPassword used for active-standby), the public SDK does
//     NOT base64-encode the password for ModifyUMemPassword; the API requires
//     the encoded form, so encode it here.
func modifyDistributedRedisPassword(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn

	req := conn.NewModifyUMemPasswordRequest()
	req.SpaceId = ucloud.String(d.Id())
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

	_, err := conn.ModifyUMemPassword(req)
	if err != nil {
		return fmt.Errorf("error on %s to redis instance %q, %s", "ModifyUMemPassword", d.Id(), err)
	}
	d.SetPartial("password")

	if err := client.waitDistributedRedisRunning(d.Id()); err != nil {
		return fmt.Errorf("error on waiting for %s complete to redis instance %q, %s", "ModifyUMemPassword", d.Id(), err)
	}
	return nil
}

func readActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading redis instance, %s", err)
	}

	inst, err := client.describeActiveStandbyRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading redis instance %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", inst.Zone)
	d.Set("name", inst.Name)
	d.Set("tag", inst.Tag)
	d.Set("charge_type", upperCamelCvt.convert(inst.ChargeType))
	d.Set("instance_type", fmt.Sprintf("redis-master-%v", inst.Size))
	d.Set("vpc_id", inst.VPCId)
	d.Set("subnet_id", inst.SubnetId)
	d.Set("engine_version", inst.Version)

	d.Set("ip_set", []map[string]interface{}{{
		"ip":   inst.VirtualIP,
		"port": inst.Port,
	}})
	d.Set("standby_zone", inst.SlaveZone)
	d.Set("auto_backup", inst.AutoBackup)
	d.Set("backup_begin_time", inst.BackupTime)
	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	d.Set("status", inst.State)

	// block_set / proxy_set / read_mode are distributed-redis-only attributes on
	// the shared schema. Set them empty for active-standby instances: a computed
	// attribute absent from state is marked <computed> in every subsequent plan,
	// producing a perpetual diff.
	d.Set("block_set", []map[string]interface{}{})
	d.Set("proxy_set", []map[string]interface{}{})
	d.Set("read_mode", "")
	return nil
}

func readDistributedRedisInstance(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading redis instance, %s", err)
	}

	inst, err := client.describeDistributedRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading redis instance %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", inst.Zone)
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

	// block_set / read_mode / proxy_set are auxiliary computed attributes.
	// DescribeUMemBlockInfo / DescribeUDRedisProxyInfo are not accessible for
	// every account/region (observed: RetCode 200 "Get names errorget access
	// failed"). Never fail the whole read over them: a failing read also breaks
	// the refresh that `terraform destroy` performs, which would strand a
	// created instance. Log a warning and fall back to empty values.
	//
	// The attributes are ALWAYS written, even when empty: a computed attribute
	// absent from state is marked <computed> in every subsequent plan, which
	// produces a perpetual diff (the acceptance framework rejects that with
	// "the plan was not empty").
	blocks, readMode, err := client.describeDistributedRedisBlockInfoById(d.Id(), inst.Zone)
	if err != nil {
		log.Printf("[WARN] could not read block info of redis instance %q: %s", d.Id(), err)
		blocks, readMode = nil, ""
	}

	blockSet := []map[string]interface{}{}
	for _, block := range blocks {
		blockSet = append(blockSet, map[string]interface{}{
			"block_id":          block.BlockId,
			"block_name":        block.BlockName,
			"block_port":        block.BlockPort,
			"block_read_weight": block.BlockReadWeight,
			"block_size":        block.BlockSize,
			"block_slot_begin":  block.BlockSlotBegin,
			"block_slot_end":    block.BlockSlotEnd,
			"block_state":       block.BlockState,
			"block_type":        block.BlockType,
			"block_used_size":   block.BlockUsedSize,
			"block_vip":         block.BlockVip,
		})
	}

	if err := d.Set("block_set", blockSet); err != nil {
		return err
	}
	d.Set("read_mode", readMode)

	proxies, err := client.describeDistributedRedisProxyInfoById(d.Id(), inst.Zone)
	if err != nil {
		log.Printf("[WARN] could not read proxy info of redis instance %q: %s", d.Id(), err)
		proxies = nil
	}

	proxySet := []map[string]interface{}{}
	for _, proxy := range proxies {
		proxySet = append(proxySet, map[string]interface{}{
			"proxy_id":    proxy.ProxyId,
			"resource_id": proxy.ResourceId,
			"state":       proxy.State,
			"vip":         proxy.Vip,
		})
	}

	if err := d.Set("proxy_set", proxySet); err != nil {
		return err
	}

	d.Set("create_time", timestampToString(inst.CreateTime))
	d.Set("expire_time", timestampToString(inst.ExpireTime))
	d.Set("status", inst.State)
	return nil
}

func deleteActiveStandbyRedisInstance(d *schema.ResourceData, meta interface{}) *resource.RetryError {
	client, err := clientFromMeta(meta)
	if err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on getting client when deleting redis instance, %s", err))
	}
	conn := client.umemconn

	req := conn.NewDeleteURedisGroupRequest()
	req.GroupId = ucloud.String(d.Id())

	if _, err := conn.DeleteURedisGroup(req); err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on deleting redis instance %q, %s", d.Id(), err))
	}

	_, err = client.describeActiveStandbyRedisById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return resource.NonRetryableError(fmt.Errorf("error on reading redis instance when deleting %q, %s", d.Id(), err))
	}

	return resource.RetryableError(fmt.Errorf("the specified redis instance %q has not been deleted due to unknown error", d.Id()))
}

func deleteDistributedRedisInstance(d *schema.ResourceData, meta interface{}) *resource.RetryError {
	client, err := clientFromMeta(meta)
	if err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on getting client when deleting redis instance, %s", err))
	}
	conn := client.umemconn

	req := conn.NewDeleteUMemSpaceRequest()
	req.SpaceId = ucloud.String(d.Id())
	// DeleteUMemSpace validates the zone common param on the backend
	// ("Missing Params [zone_id]" without it); availability_zone is kept in
	// state by the read path.
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	if _, err := conn.DeleteUMemSpace(req); err != nil {
		return resource.NonRetryableError(fmt.Errorf("error on deleting redis instance %q, %s", d.Id(), err))
	}

	_, err = client.describeDistributedRedisById(d.Id())
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

func getRedisDefaultParameterGroup(d *schema.ResourceData, client *productClient) (string, error) {
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

func (c *productClient) waitActiveStandbyRedisRunning(id string) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeActiveStandbyRedisById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		if resp.State != statusRunning {
			return resp, statusPending, nil
		}
		return resp, statusInitialized, nil
	}

	return waitForMemoryInstance(refresh, memoryInstanceWaitTimeout)
}

// waitActiveStandbyRedisTransform blocks until an ISolationURedisGroup transform
// reaches its target state: "Bind" ends in Running, "UNBind" ends in ISolation.
// The transform is async and the instance still reports the OLD state right
// after the API returns, so a wait accepting both states would return
// immediately and the next operation (e.g. a restart after Bind) would run
// against an instance that has not come back up yet.
func (c *productClient) waitActiveStandbyRedisTransform(id, transformType string) error {
	targetState := statusUNBind
	if transformType == "Bind" {
		targetState = statusRunning
	}

	refresh := func() (interface{}, string, error) {
		resp, err := c.describeActiveStandbyRedisById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		if resp.State != targetState {
			return resp, statusPending, nil
		}
		return resp, statusInitialized, nil
	}

	return waitForMemoryInstance(refresh, memoryInstanceWaitTimeout)
}

func (c *productClient) waitDistributedRedisRunning(id string) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeDistributedRedisById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		if resp.State != statusRunning {
			return resp, statusPending, nil
		}
		return resp, statusInitialized, nil
	}

	return waitForMemoryInstance(refresh, memoryInstanceWaitTimeout)
}

// waitActiveStandbyRedisResized blocks until the instance is Running AND reports
// the target size in GB. Waiting on "Running" alone races with the async resize:
// right after ResizeURedisGroup returns, the instance can still show the old size
// with state Running before the backend flips it to Resizing, so a Running-only
// wait returns early and the subsequent read persists the stale size.
func (c *productClient) waitActiveStandbyRedisResized(id string, size int) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeActiveStandbyRedisById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		// surface a failed resize immediately instead of polling until timeout
		if resp.State == statusResizeFail || resp.State == statusFail {
			return nil, "", fmt.Errorf("resize of redis instance %q failed, instance state: %q", id, resp.State)
		}

		if resp.State != statusRunning || resp.Size != size {
			return resp, statusPending, nil
		}
		return resp, statusInitialized, nil
	}

	return waitForMemoryInstance(refresh, redisResizeWaitTimeout)
}

// waitDistributedRedisResized blocks until the instance is Running AND reports
// the target size in GB; see waitActiveStandbyRedisResized for why a Running-only
// wait is not enough after ResizeUMemSpace.
func (c *productClient) waitDistributedRedisResized(id string, size int) error {
	refresh := func() (interface{}, string, error) {
		resp, err := c.describeDistributedRedisById(id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		// surface a failed resize immediately instead of polling until timeout
		if resp.State == statusResizeFail || resp.State == statusFail {
			return nil, "", fmt.Errorf("resize of redis instance %q failed, instance state: %q", id, resp.State)
		}

		if resp.State != statusRunning || resp.Size != size {
			return resp, statusPending, nil
		}
		return resp, statusInitialized, nil
	}

	return waitForMemoryInstance(refresh, redisResizeWaitTimeout)
}

// udredisBlockSizes are the per-shard capacities (GB) accepted by
// ResizeUDRedisBlockSize.
var udredisBlockSizes = []int{4, 8, 12, 16, 20}

// resizeDistributedRedisInstance grows a distributed redis (UDRedis) instance.
//
// UDRedis capacity is shardCount x per-shard size, and resizing happens per
// shard through ResizeUDRedisBlockSize (BlockId required, per-shard capacity
// restricted to 4/8/12/16/20 GB). ResizeUMemSpace must NOT be used here: the
// backend accepts it with RetCode 0 but never starts a resize for a UDRedis
// instance - no shard enters Resizing and the total size never changes.
//
// Shards are resized one by one and the loop is idempotent (a shard already at
// the target size is skipped), so a partially failed resize can simply be
// retried on the next apply.
func resizeDistributedRedisInstance(client *productClient, d *schema.ResourceData) error {
	conn := client.umemconn
	id := d.Id()
	zone := d.Get("availability_zone").(string)
	targetTotal := getRedisCapability(d.Get("instance_type").(string))

	blocks, _, err := client.describeDistributedRedisBlockInfoById(id, zone)
	if err != nil {
		return fmt.Errorf("error on listing shards of redis instance %q before resizing (distributed resize requires DescribeUMemBlockInfo access): %s", id, err)
	}
	if len(blocks) == 0 {
		return fmt.Errorf("cannot resize redis instance %q: no shard information returned by DescribeUMemBlockInfo", id)
	}

	if targetTotal%len(blocks) != 0 {
		return fmt.Errorf("cannot resize redis instance %q: total size %dGB is not divisible by its %d shards", id, targetTotal, len(blocks))
	}
	targetBlockSize := targetTotal / len(blocks)
	validSize := false
	for _, size := range udredisBlockSizes {
		if size == targetBlockSize {
			validSize = true
			break
		}
	}
	if !validSize {
		return fmt.Errorf("cannot resize redis instance %q: per-shard size %dGB is invalid, must be one of 4/8/12/16/20GB", id, targetBlockSize)
	}

	for _, block := range blocks {
		if block.BlockSize == targetBlockSize {
			continue // already at target, e.g. retrying a partially failed resize
		}

		req := conn.NewResizeUDRedisBlockSizeRequest()
		req.SpaceId = ucloud.String(id)
		req.Zone = ucloud.String(zone)
		req.BlockId = ucloud.String(block.BlockId)
		req.BlockSize = ucloud.Int(targetBlockSize)

		if _, err := conn.ResizeUDRedisBlockSize(req); err != nil {
			return fmt.Errorf("error on resizing shard %q of redis instance %q, %s", block.BlockId, id, err)
		}

		if err := client.waitDistributedRedisBlockResized(id, zone, block.BlockId, targetBlockSize); err != nil {
			return fmt.Errorf("error on waiting for shard %q of redis instance %q complete resizing, %s", block.BlockId, id, err)
		}
	}

	// final space-level consistency wait: Running and the total size reached
	return client.waitDistributedRedisResized(id, targetTotal)
}

// waitDistributedRedisBlockResized blocks until the given shard reports Running
// with the target per-shard size, failing fast on a terminal shard failure state.
func (c *productClient) waitDistributedRedisBlockResized(id, zone, blockId string, size int) error {
	refresh := func() (interface{}, string, error) {
		blocks, _, err := c.describeDistributedRedisBlockInfoById(id, zone)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}
		for _, block := range blocks {
			if block.BlockId != blockId {
				continue
			}
			if block.BlockState == statusResizeFail || block.BlockState == statusFail {
				return nil, "", fmt.Errorf("resize of shard %q failed, shard state: %q", blockId, block.BlockState)
			}
			if block.BlockState == statusRunning && block.BlockSize == size {
				return block, statusInitialized, nil
			}
			return block, statusPending, nil
		}
		return nil, "", fmt.Errorf("shard %q of redis instance %q disappeared during resize", blockId, id)
	}

	return waitForMemoryInstance(refresh, redisResizeWaitTimeout)
}

// diffIsDestroyPlan reports whether the diff was produced for a destroy plan.
// SDKv1 still runs CustomizeDiff while planning a destroy with an empty config
// (schemaMap.Diff only skips it for DestroyTainted), where required fields read
// back as nil/zero values: type assertions panic and GetOk/GetOkExists disagree,
// which previously surfaced as spurious errors during acceptance-test cleanup.
// instance_type is Required on both redis and memcache, so an absent value
// means there is no new configuration to validate.
func diffIsDestroyPlan(diff *schema.ResourceDiff) bool {
	v, _ := diff.Get("instance_type").(string)
	return v == ""
}

func diffValidateRedisInstanceType(old, new, meta interface{}) error {
	oldValue, _ := old.(string)
	newValue, _ := new.(string)
	if newValue == "" {
		// destroy plan: no new configuration to validate
		return nil
	}
	if len(oldValue) > 0 {
		oldType, _ := parseRedisInstanceType(oldValue)
		newType, _ := parseRedisInstanceType(newValue)
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
	if diffIsDestroyPlan(diff) {
		return nil
	}

	engineVersion, _ := diff.Get("engine_version").(string)
	instanceType, _ := diff.Get("instance_type").(string)
	redisType, _ := parseRedisInstanceType(instanceType)

	if redisType.Type == "master" && engineVersion == "" {
		return fmt.Errorf("the %q argument must be set to active-standby redis instance", "engine_version")
	}

	if redisType.Type == "distributed" && engineVersion != "" {
		return fmt.Errorf("the %q argument is not apply to distributed redis instance", "engine_version")
	}

	return nil
}

func diffValidateRedisStandbyZone(diff *schema.ResourceDiff, v interface{}) error {
	if diffIsDestroyPlan(diff) {
		return nil
	}

	zone, _ := diff.Get("availability_zone").(string)

	if val, ok := diff.GetOk("standby_zone"); ok && val.(string) == zone {
		return fmt.Errorf("standby_zone %q must be different from availability_zone %q", val.(string), zone)
	}

	return nil
}

func diffValidateBackup(diff *schema.ResourceDiff, meta interface{}) error {
	if diffIsDestroyPlan(diff) {
		return nil
	}

	_, okT := diff.GetOkExists("backup_begin_time")
	_, okA := diff.GetOk("auto_backup")
	instanceType, _ := diff.Get("instance_type").(string)
	t, err := parseRedisInstanceType(instanceType)
	if err != nil {
		return err
	}
	if t.Type == "distributed" && (okT || okA) {
		return fmt.Errorf("the distributed redis not support backup in terraform, please use console or else if you want to use")
	}

	if !okA && okT {
		return fmt.Errorf("the argument %q is required when set %q", "auto_backup", "backup_begin_time")
	}

	return nil
}

// diffValidateBlockCnt rejects block_cnt on active-standby redis: sharding only
// applies to the distributed instance type (CreateUMemSpace), not to the
// active-standby URedisGroup path.
func diffValidateBlockCnt(diff *schema.ResourceDiff, meta interface{}) error {
	if diffIsDestroyPlan(diff) {
		return nil
	}

	instanceType, _ := diff.Get("instance_type").(string)
	t, err := parseRedisInstanceType(instanceType)
	if err != nil {
		return err
	}

	if _, ok := diff.GetOk("block_cnt"); ok && t.Type != "distributed" {
		return fmt.Errorf("the argument %q is only supported for distributed redis, not supported for active-standby redis", "block_cnt")
	}

	return nil
}

// diffValidateRedisRestart rejects restart_trigger on distributed redis: the SDK only
// exposes RestartURedisGroup for active-standby redis, with no restart API for distributed.
func diffValidateRedisRestart(diff *schema.ResourceDiff, meta interface{}) error {
	if diffIsDestroyPlan(diff) {
		return nil
	}

	instanceType, _ := diff.Get("instance_type").(string)
	t, err := parseRedisInstanceType(instanceType)
	if err != nil {
		return err
	}

	if _, ok := diff.GetOk("restart_trigger"); ok && t.Type == "distributed" {
		return fmt.Errorf("the argument %q is only supported for active-standby redis, not supported for distributed redis", "restart_trigger")
	}

	return nil
}
