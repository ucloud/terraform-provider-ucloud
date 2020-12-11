package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
		CustomizeDiff: customdiff.All(
			diffValidateInternalWithSubnetId,
		),

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
				ForceNew: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"listen_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"request_proxy",
					"packets_transmit",
				}, false),
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

	if v, ok := d.GetOk("listen_type"); ok {
		req.ListenType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	}

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

	var internal bool
	if val, ok := d.GetOk("internal"); ok {
		internal = val.(bool)
		if internal {
			req.InnerMode = ucloud.String("Yes")
		} else {
			req.OuterMode = ucloud.String("Yes")
		}
	} else {
		req.OuterMode = ucloud.String("Yes")
	}

	if val, ok := d.GetOk("security_group"); ok {
		if internal && val != "" {
			return fmt.Errorf("the security_group only takes effect for ULB instances of request_proxy mode and extranet mode at present, got internal = %t", internal)
		}
		req.FirewallId = ucloud.String(val.(string))
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
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	d.Partial(true)

	if d.HasChange("security_group") && !d.IsNewResource() {
		if val, ok := d.GetOk("internal"); ok {
			if val.(bool) {
				return fmt.Errorf("the security_group only takes effect for ULB instances of request_proxy mode and extranet mode at present, got internal = %t", val.(bool))
			}
		}
		conn := client.unetconn
		req := conn.NewGrantFirewallRequest()
		req.FWId = ucloud.String(d.Get("security_group").(string))
		req.ResourceType = ucloud.String(eipResourceTypeULB)
		req.ResourceId = ucloud.String(d.Id())

		_, err := conn.GrantFirewall(req)
		if err != nil {
			return fmt.Errorf("error on %s to lb %q, %s", "GrantFirewall", d.Id(), err)
		}

		d.SetPartial("security_group")
	}

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

	listenType := upperCamelCvt.convert(lbSet.ListenType)
	if listenType == "request_proxy" || listenType == "packets_transmit" {
		d.Set("listen_type", listenType)
	}

	d.Set("name", lbSet.Name)
	d.Set("tag", lbSet.Tag)
	d.Set("remark", lbSet.Remark)
	d.Set("create_time", timestampToString(lbSet.CreateTime))
	d.Set("vpc_id", lbSet.VPCId)
	d.Set("private_ip", lbSet.PrivateIP)

	if notEmptyStringInSet(lbSet.SubnetId) {
		d.Set("subnet_id", lbSet.SubnetId)
	}

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

	sgSet, err := client.describeFirewallByIdAndType(d.Id(), eipResourceTypeULB)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error on reading security group when reading lb %q, %s", d.Id(), err)
	}

	d.Set("security_group", sgSet.FWId)

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
		Timeout:    3 * time.Minute,
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

func diffValidateInternalWithSubnetId(diff *schema.ResourceDiff, meta interface{}) error {
	var internal bool
	var subnetId string

	if v, ok := diff.GetOk("internal"); ok {
		internal = v.(bool)
	}
	if v, ok := diff.GetOk("subnet_id"); ok {
		subnetId = v.(string)
	}

	if !internal && subnetId != "" {
		return fmt.Errorf("the lb instance cannot set %q, When the %q is true", "subnet_id", "internal")
	}

	return nil
}
