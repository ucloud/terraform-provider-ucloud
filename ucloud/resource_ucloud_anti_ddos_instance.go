package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudAntiDDoSInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudAntiDDoSInstanceCreate,
		Read:   resourceUCloudAntiDDoSInstanceRead,
		Update: resourceUCloudAntiDDoSInstanceUpdate,
		Delete: resourceUCloudAntiDDoSInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAntiDDoSInstanceName,
			},

			"area": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"EastChina",
					"NorthChina",
				}, false),
			},

			"data_center": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Zaozhuang",
					"Yangzhou",
					"Taizhou",
					"Shijiazhuang",
				}, false),
			},

			"bandwidth": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(50),
			},

			"base_defence_value": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(30),
			},

			"max_defence_value": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(30),
			},
			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"month",
					"year",
				}, false),
			},
			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
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

func resourceUCloudAntiDDoSInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	err := validateAntiDDoSInstance(d)
	if err != nil {
		return err
	}
	client := meta.(*UCloudClient)
	conn := client.uadsconn

	req := conn.NewBuyHighProtectGameServiceRequest()
	req.HighProtectGameServiceName = ucloud.String(d.Get("name").(string))
	req.AreaLine = ucloud.String(d.Get("area").(string))
	req.EngineRoom = []string{d.Get("data_center").(string)}
	req.SrcBandwidth = ucloud.Int(d.Get("bandwidth").(int))
	req.DefenceDDosBaseFlowArr = []string{strconv.Itoa(d.Get("base_defence_value").(int))}
	req.DefenceDDosMaxFlowArr = []string{strconv.Itoa(d.Get("max_defence_value").(int))}

	req.LineType = ucloud.String("BGP")
	req.DefenceType = ucloud.String("TypeDynamic")

	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}
	resp, err := conn.BuyHighProtectGameService(req)
	if err != nil {
		return fmt.Errorf("error on creating ucloud_anti_ddos_instance, %s", err)
	}

	d.SetId(resp.ResourceInfo.ResourceId)

	// after create lb, we need to wait it initialized
	stateConf := antiDDoSInstanceWaitForState(client, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for ucloud_anti_ddos_instance %q creating, %s", d.Id(), err)
	}

	return resourceUCloudAntiDDoSInstanceRead(d, meta)
}

func resourceUCloudAntiDDoSInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn
	d.Partial(true)
	if err := validateAntiDDoSInstance(d); err != nil {
		return err
	}
	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyHighProtectGameServiceRequest()
		req.ResourceId = ucloud.String(d.Id())
		req.HighProtectGameServiceName = ucloud.String(d.Get("name").(string))
		_, err := conn.ModifyHighProtectGameService(req)
		if err != nil {
			return fmt.Errorf("error on %s to update name, %s", d.Id(), err)
		}
		d.SetPartial("name")
	}

	if (d.HasChange("bandwidth") || d.HasChange("base_defence_value") || d.HasChange("max_defence_value")) && !d.IsNewResource() {
		req := conn.NewUpgradeHighProtectGameServiceRequest()
		req.ResourceId = ucloud.String(d.Id())
		req.AreaLine = ucloud.String(d.Get("area").(string))
		req.EngineRoom = []string{d.Get("data_center").(string)}
		req.SrcBandwidth = ucloud.Int(d.Get("bandwidth").(int))
		req.DefenceDDosBaseFlowArr = []string{strconv.Itoa(d.Get("base_defence_value").(int))}
		req.DefenceDDosMaxFlowArr = []string{strconv.Itoa(d.Get("max_defence_value").(int))}

		req.LineType = ucloud.String("BGP")
		req.DefenceType = ucloud.String("TypeDynamic")
		_, err := conn.UpgradeHighProtectGameService(req)
		if err != nil {
			return fmt.Errorf("error on %s to update ucloud_anti_ddos_instance, %s", d.Id(), err)
		}
		d.SetPartial("bandwidth")
		d.SetPartial("base_defence_value")
		d.SetPartial("max_defence_value")
	}

	d.Partial(false)

	return resourceUCloudAntiDDoSInstanceRead(d, meta)
}

func resourceUCloudAntiDDoSInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	uadsServiceInfo, err := client.describeUADSById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading uads %q, %s", d.Id(), err)
	}
	if len(uadsServiceInfo.EngineRoom) == 0 {
		return fmt.Errorf("fail to get data center info, %s", d.Id())
	}
	d.Set("name", uadsServiceInfo.Name)
	d.Set("area", uadsServiceInfo.AreaLine)
	d.Set("data_center", uadsServiceInfo.EngineRoom[0])
	d.Set("bandwidth", uadsServiceInfo.SrcBandwidth)
	d.Set("base_defence_value", uadsServiceInfo.DefenceDDosBaseFlowArr[0])
	d.Set("max_defence_value", uadsServiceInfo.DefenceDDosMaxFlowArr[0])
	d.Set("status", uadsServiceInfo.DefenceStatus)
	d.Set("create_time", timestampToString(uadsServiceInfo.CreateTime))
	d.Set("expire_time", timestampToString(uadsServiceInfo.ExpiredTime))

	return nil
}

func resourceUCloudAntiDDoSInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uadsconn

	req := conn.NewDeleteHighProtectGameServiceRequest()
	req.ResourceId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteHighProtectGameService(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting ucloud_anti_ddos_instance %q, %s", d.Id(), err))
		}

		_, err := client.describeUADSById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading ucloud_anti_ddos_instance when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified ucloud_anti_ddos_instance %q has not been deleted due to unknown error", d.Id()))
	})
}

func antiDDoSInstanceWaitForState(client *UCloudClient, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			uadsInfo, err := client.describeUADSById(id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			return uadsInfo, statusInitialized, nil
		},
	}
}

func validateAntiDDoSInstance(d *schema.ResourceData) error {
	baseDefenceValue := d.Get("base_defence_value").(int)
	maxDefenceValue := d.Get("max_defence_value").(int)
	if maxDefenceValue < baseDefenceValue {
		return fmt.Errorf("max_defence_value %v cannot be less than base_defence_value %v", maxDefenceValue, baseDefenceValue)
	}
	return nil
}
