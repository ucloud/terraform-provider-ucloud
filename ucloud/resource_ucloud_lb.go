package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLB() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBCreate,
		Read:   resourceUCloudLBRead,
		Update: resourceUCloudLBUpdate,
		Delete: resourceUCloudLBDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"internal": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"internet_charge_type": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "Month",
				Optional:     true,
				ValidateFunc: validateStringInChoices([]string{"Month", "Year", "Dynamic"}),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "LB",
			},

			"tag": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"remark": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip_set": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"internet_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"ip": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"eip_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"private_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudLBCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn

	req := conn.NewCreateULBRequest()
	req.ChargeType = ucloud.String(d.Get("internet_charge_type").(string))
	req.ULBName = ucloud.String(d.Get("name").(string))

	if val, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(val.(string))
	}

	if d.Get("internal").(bool) {
		req.InnerMode = ucloud.String("Yes")
	} else {
		req.OuterMode = ucloud.String("Yes")
	}

	resp, err := conn.CreateULB(req)

	if err != nil {
		return fmt.Errorf("error in create lb, %s", err)
	}

	d.SetId(resp.ULBId)

	time.Sleep(5 * time.Second)

	return resourceUCloudLBUpdate(d, meta)
}

func resourceUCloudLBUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn

	d.Partial(true)

	isChanged := false
	req := conn.NewUpdateULBAttributeRequest()
	req.ULBId = ucloud.String(d.Id())

	if d.HasChange("name") && !d.IsNewResource() {
		isChanged = true
		req.Name = ucloud.String(d.Get("name").(string))
		d.SetPartial("name")
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("tag").(string))
		d.SetPartial("tag")
	}

	if d.HasChange("remark") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("remark").(string))
		d.SetPartial("remark")
	}

	if isChanged {
		_, err := conn.UpdateULBAttribute(req)

		if err != nil {
			return fmt.Errorf("do %s failed in update lb %s, %s", "UpdateULBAttribute", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudLBRead(d, meta)
}

func resourceUCloudLBRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	lbSet, err := client.describeLBById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read lb %s, %s", "DescribeULB", d.Id(), err)
	}

	d.Set("name", lbSet.Name)
	d.Set("tag", lbSet.Tag)
	d.Set("remark", lbSet.Remark)
	d.Set("create_time", timestampToString(lbSet.CreateTime))
	d.Set("expire_time", timestampToString(lbSet.ExpireTime))
	d.Set("vpc_id", lbSet.VPCId)
	d.Set("subnet_id", lbSet.SubnetId)

	//TODO: [API-ERROR]need ulbSet.ChargeType
	d.Set("internet_charge_type", d.Get("internet_charge_type").(string))
	d.Set("private_ip", lbSet.PrivateIP)

	ipSet := []map[string]interface{}{}
	for _, item := range lbSet.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"internet_type": item.OperatorName,
			"ip":            item.EIP,
			"eip_id":        item.EIPId,
		})
	}
	d.Set("ip_set", ipSet)

	return nil
}

func resourceUCloudLBDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	req := conn.NewDeleteULBRequest()
	req.ULBId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteULB(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete lb %s, %s", d.Id(), err))
		}

		_, err := client.describeLBById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete lb %s, %s", "DescribeULB", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete lb but it still exists"))
	})
}
