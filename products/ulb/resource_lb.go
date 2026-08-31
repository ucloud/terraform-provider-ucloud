package ulb

import (
	"fmt"
	"time"

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

func resourceUCloudLBCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb, %s", err)
	}
	conn := client.ulbconn

	request := conn.NewCreateULBRequest()
	if value, ok := data.GetOk("listen_type"); ok {
		request.ListenType = ucloud.String(upperCamelUnconvert(value.(string)))
	}
	if value, ok := data.GetOk("name"); ok {
		request.ULBName = ucloud.String(value.(string))
	} else {
		request.ULBName = ucloud.String(resource.PrefixedUniqueId("tf-lb-"))
	}
	if value, ok := data.GetOk("tag"); ok {
		request.Tag = ucloud.String(value.(string))
	} else {
		request.Tag = ucloud.String(defaultTag)
	}
	if value, ok := data.GetOk("remark"); ok {
		request.Remark = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("vpc_id"); ok {
		request.VPCId = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("subnet_id"); ok {
		request.SubnetId = ucloud.String(value.(string))
	}

	var internal bool
	if value, ok := data.GetOk("internal"); ok {
		internal = value.(bool)
		if internal {
			request.InnerMode = ucloud.String("Yes")
		} else {
			request.OuterMode = ucloud.String("Yes")
		}
	} else {
		request.OuterMode = ucloud.String("Yes")
	}
	if value, ok := data.GetOk("security_group"); ok {
		if internal && value != "" {
			return fmt.Errorf("the security_group only takes effect for ULB instances of request_proxy mode and extranet mode at present, got internal = %t", internal)
		}
		request.FirewallId = ucloud.String(value.(string))
	}

	response, err := conn.CreateULB(request)
	if err != nil {
		return fmt.Errorf("error on creating lb, %s", err)
	}
	data.SetId(response.ULBId)

	stateConf := lbWaitForState(client, data.Id())
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for lb %q complete creating, %s", data.Id(), err)
	}
	return resourceUCloudLBRead(data, meta)
}

func resourceUCloudLBUpdate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating lb, %s", err)
	}
	conn := client.ulbconn

	data.Partial(true)
	if data.HasChange("security_group") && !data.IsNewResource() {
		if value, ok := data.GetOk("internal"); ok && value.(bool) {
			return fmt.Errorf("the security_group only takes effect for ULB instances of request_proxy mode and extranet mode at present, got internal = %t", value.(bool))
		}
		request := client.unetconn.NewGrantFirewallRequest()
		request.FWId = ucloud.String(data.Get("security_group").(string))
		request.ResourceType = ucloud.String(eipResourceTypeULB)
		request.ResourceId = ucloud.String(data.Id())
		if _, err := client.unetconn.GrantFirewall(request); err != nil {
			return fmt.Errorf("error on %s to lb %q, %s", "GrantFirewall", data.Id(), err)
		}
		data.SetPartial("security_group")
	}

	isChanged := false
	request := conn.NewUpdateULBAttributeRequest()
	request.ULBId = ucloud.String(data.Id())
	if data.HasChange("name") && !data.IsNewResource() {
		isChanged = true
		request.Name = ucloud.String(data.Get("name").(string))
	}
	if data.HasChange("tag") && !data.IsNewResource() {
		isChanged = true
		if value, ok := data.GetOk("tag"); ok {
			request.Tag = ucloud.String(value.(string))
		} else {
			request.Tag = ucloud.String(defaultTag)
		}
	}
	if data.HasChange("remark") && !data.IsNewResource() {
		isChanged = true
		// Preserve the legacy request field behavior for state compatibility.
		request.Tag = ucloud.String(data.Get("remark").(string))
	}
	if isChanged {
		if _, err := conn.UpdateULBAttribute(request); err != nil {
			return fmt.Errorf("error on %s to lb %q, %s", "UpdateULBAttribute", data.Id(), err)
		}
		data.SetPartial("name")
		data.SetPartial("tag")
		data.SetPartial("remark")
	}
	data.Partial(false)
	return resourceUCloudLBRead(data, meta)
}

func resourceUCloudLBRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb, %s", err)
	}

	lbSet, err := client.describeLBById(data.Id())
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading lb %q, %s", data.Id(), err)
	}

	listenType := upperCamelConvert(lbSet.ListenType)
	if listenType == "request_proxy" || listenType == "packets_transmit" {
		_ = data.Set("listen_type", listenType)
	}
	_ = data.Set("name", lbSet.Name)
	_ = data.Set("tag", lbSet.Tag)
	_ = data.Set("remark", lbSet.Remark)
	_ = data.Set("create_time", timestampToString(lbSet.CreateTime))
	_ = data.Set("vpc_id", lbSet.VPCId)
	_ = data.Set("private_ip", lbSet.PrivateIP)
	if notEmptyStringInSet(lbSet.SubnetId) {
		_ = data.Set("subnet_id", lbSet.SubnetId)
	}
	if lbSet.ULBType == "OuterMode" {
		_ = data.Set("internal", false)
	} else if lbSet.ULBType == "InnerMode" {
		_ = data.Set("internal", true)
	}

	ipSet := []map[string]interface{}{}
	for _, item := range lbSet.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"internet_type": item.OperatorName,
			"ip":            item.EIP,
		})
	}
	if err := data.Set("ip_set", ipSet); err != nil {
		return err
	}

	securityGroup, err := client.describeFirewallByIdAndType(data.Id(), eipResourceTypeULB)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error on reading security group when reading lb %q, %s", data.Id(), err)
	}
	_ = data.Set("security_group", securityGroup.FWId)
	return nil
}

func resourceUCloudLBDelete(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb, %s", err)
	}
	request := client.ulbconn.NewDeleteULBRequest()
	request.ULBId = ucloud.String(data.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.ulbconn.DeleteULB(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb %q, %s", data.Id(), err))
		}
		_, err := client.describeLBById(data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb when deleting %q, %s", data.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified lb %q has not been deleted due to unknown error", data.Id()))
	})
}

func lbWaitForState(client *productClient, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			lbSet, err := client.describeLBById(id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return lbSet, statusInitialized, nil
		},
	}
}
