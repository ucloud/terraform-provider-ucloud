package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudNetworkInterfaceCreate,
		Update: resourceUCloudNetworkInterfaceUpdate,
		Read:   resourceUCloudNetworkInterfaceRead,
		Delete: resourceUCloudNetworkInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
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
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating network interface, %s", err)
	}
	conn := client.vpcconn

	req := conn.NewCreateNetworkInterfaceRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))

	var name string
	if value, ok := d.GetOk("name"); ok {
		name = value.(string)
	} else {
		name = resource.PrefixedUniqueId("tf-uni-")
	}
	req.Name = ucloud.String(name)
	req.Tag = ucloud.String(d.Get("tag").(string))

	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("private_ip"); ok {
		req.PrivateIp = []string{value.(string)}
	}

	resp, err := conn.CreateNetworkInterface(req)
	if err != nil {
		return fmt.Errorf("error on creating network interface, %s", err)
	}
	d.SetId(resp.NetworkInterface.InterfaceId)
	// Persist the generated/configured name here so state stays consistent even
	// when the read path returns a slightly different value.
	_ = d.Set("name", name)

	if value, ok := d.GetOk("instance_id"); ok && value.(string) != "" {
		attachReq := conn.NewAttachNetworkInterfaceRequest()
		attachReq.InstanceId = ucloud.String(value.(string))
		attachReq.InterfaceId = ucloud.String(d.Id())
		if _, err := conn.AttachNetworkInterface(attachReq); err != nil {
			return fmt.Errorf("error on attaching network interface %q to instance %q, %s", d.Id(), value.(string), err)
		}
	}

	return resourceUCloudNetworkInterfaceRead(d, meta)
}

func resourceUCloudNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating network interface, %s", err)
	}
	conn := client.vpcconn

	d.Partial(true)

	if d.HasChange("name") || d.HasChange("tag") || d.HasChange("remark") {
		if err := client.modifyNetworkInterfaceAttribute(
			d.Id(),
			d.Get("name").(string),
			d.Get("tag").(string),
			d.Get("remark").(string),
		); err != nil {
			return fmt.Errorf("error on modifying network interface %q attribute, %s", d.Id(), err)
		}
		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")
	}

	if d.HasChange("instance_id") && !d.IsNewResource() {
		oldValue, newValue := d.GetChange("instance_id")
		oldInstanceID := oldValue.(string)
		newInstanceID := newValue.(string)
		if oldInstanceID != "" {
			detachReq := conn.NewDetachNetworkInterfaceRequest()
			detachReq.InstanceId = ucloud.String(oldInstanceID)
			detachReq.InterfaceId = ucloud.String(d.Id())
			if _, err := conn.DetachNetworkInterface(detachReq); err != nil {
				return fmt.Errorf("error on detaching network interface %q from instance %q, %s", d.Id(), oldInstanceID, err)
			}
		}
		if newInstanceID != "" {
			attachReq := conn.NewAttachNetworkInterfaceRequest()
			attachReq.InstanceId = ucloud.String(newInstanceID)
			attachReq.InterfaceId = ucloud.String(d.Id())
			if _, err := conn.AttachNetworkInterface(attachReq); err != nil {
				return fmt.Errorf("error on attaching network interface %q to instance %q, %s", d.Id(), newInstanceID, err)
			}
		}
		d.SetPartial("instance_id")
	}

	d.Partial(false)
	return resourceUCloudNetworkInterfaceRead(d, meta)
}

func resourceUCloudNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading network interface, %s", err)
	}
	uni, err := client.describeNetworkInterfaceById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading network interface %q, %s", d.Id(), err)
	}

	_ = d.Set("vpc_id", uni.VPCId)
	_ = d.Set("subnet_id", uni.SubnetId)
	_ = d.Set("name", uni.Name)
	_ = d.Set("tag", uni.Tag)
	_ = d.Set("remark", uni.Remark)
	_ = d.Set("instance_id", uni.AttachInstanceId)
	_ = d.Set("mac_address", uni.MacAddress)
	_ = d.Set("status", uni.Status)
	_ = d.Set("create_time", timestampToString(uni.CreateTime))
	if len(uni.PrivateIpSet) > 0 {
		_ = d.Set("private_ip", uni.PrivateIpSet[0])
	}
	return nil
}

func resourceUCloudNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting network interface, %s", err)
	}
	conn := client.vpcconn

	// A network interface must be detached before it can be deleted.
	if uni, describeErr := client.describeNetworkInterfaceById(d.Id()); describeErr == nil && uni.AttachInstanceId != "" {
		detachReq := conn.NewDetachNetworkInterfaceRequest()
		detachReq.InstanceId = ucloud.String(uni.AttachInstanceId)
		detachReq.InterfaceId = ucloud.String(d.Id())
		if _, err := conn.DetachNetworkInterface(detachReq); err != nil {
			return fmt.Errorf("error on detaching network interface %q from instance %q, %s", d.Id(), uni.AttachInstanceId, err)
		}
	}

	req := conn.NewDeleteNetworkInterfaceRequest()
	req.InterfaceId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteNetworkInterface(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting network interface %q, %s", d.Id(), err))
		}
		if _, err := client.describeNetworkInterfaceById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading network interface when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified network interface %q has not been deleted due to unknown error", d.Id()))
	})
}
