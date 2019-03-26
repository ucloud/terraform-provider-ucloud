package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"charge_type": {
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "attribute `charge_type` is deprecated for optimizing parameters",
				ValidateFunc: validation.StringInSlice([]string{
					"month",
					"year",
					"dynamic",
				}, false),
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"internet_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
				Type:       schema.TypeString,
				Deprecated: "attribute `expire_time` is deprecated for optimizing outputs",
				Computed:   true,
			},
		},
	}
}

func resourceUCloudLBCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	req := conn.NewCreateULBRequest()

	if v, ok := d.GetOk("name"); ok {
		req.ULBName = ucloud.String(v.(string))
	} else {
		req.ULBName = ucloud.String(resource.PrefixedUniqueId("tf-lb-"))
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
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

	if val, ok := d.GetOk("internal"); ok {
		if val.(bool) {
			req.InnerMode = ucloud.String("Yes")
		} else {
			req.OuterMode = ucloud.String("Yes")
		}
	} else {
		req.OuterMode = ucloud.String("Yes")
	}

	resp, err := conn.CreateULB(req)
	if err != nil {
		return fmt.Errorf("error on creating lb, %s", err)
	}

	d.SetId(resp.ULBId)

	// after create lb, we need to wait it initialized
	stateConf := lbWaitForState(client, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for lb %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudLBRead(d, meta)
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
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		isChanged = true

		// if tag is empty string, use default tag
		if v, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(v.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}
	}

	if d.HasChange("remark") && !d.IsNewResource() {
		isChanged = true
		req.Tag = ucloud.String(d.Get("remark").(string))
	}

	if isChanged {
		_, err := conn.UpdateULBAttribute(req)
		if err != nil {
			return fmt.Errorf("error on %s to lb %q, %s", "UpdateULBAttribute", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")
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
		return fmt.Errorf("error on reading lb %q, %s", d.Id(), err)
	}

	d.Set("name", lbSet.Name)
	d.Set("tag", lbSet.Tag)
	d.Set("remark", lbSet.Remark)
	d.Set("create_time", timestampToString(lbSet.CreateTime))
	d.Set("vpc_id", lbSet.VPCId)
	d.Set("subnet_id", lbSet.SubnetId)
	d.Set("private_ip", lbSet.PrivateIP)

	if lbSet.ULBType == "OuterMode" {
		d.Set("internal", false)
	} else if lbSet.ULBType == "InnerMode" {
		d.Set("internal", true)
	}

	ipSet := []map[string]interface{}{}
	for _, item := range lbSet.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"internet_type": item.OperatorName,
			"ip":            item.EIP,
		})
	}

	if err := d.Set("ip_set", ipSet); err != nil {
		return err
	}

	return nil
}

func resourceUCloudLBDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	req := conn.NewDeleteULBRequest()
	req.ULBId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteULB(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb %q, %s", d.Id(), err))
		}

		_, err := client.describeLBById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified lb %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbWaitForState(client *UCloudClient, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			eip, err := client.describeLBById(id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			return eip, statusInitialized, nil
		},
	}
}
