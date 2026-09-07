package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func resourceUCloudSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSubnetCreate,
		Update: resourceUCloudSubnetUpdate,
		Read:   resourceUCloudSubnetRead,
		Delete: resourceUCloudSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRBlock,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				ForceNew: true,
				Computed: true,
			},
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating subnet, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewCreateSubnetRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	cidrBlock := d.Get("cidr_block").(string)
	cidr, _ := parseCidrBlock(cidrBlock)
	req.Subnet = ucloud.String(cidr.Network)
	req.Netmask = ucloud.Int(cidr.Mask)
	if value, ok := d.GetOk("name"); ok {
		req.SubnetName = ucloud.String(value.(string))
	} else {
		req.SubnetName = ucloud.String(resource.PrefixedUniqueId("tf-subnet-"))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}
	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
	}
	resp, err := conn.CreateSubnet(req)
	if err != nil {
		return fmt.Errorf("error on creating subnet, %s", err)
	}
	d.SetId(resp.SubnetId)
	if _, err = subnetWaitForState(client, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for subnet %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudSubnetRead(d, meta)
}

func resourceUCloudSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating subnet, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	req := conn.NewUpdateSubnetAttributeRequest()
	req.SubnetId = ucloud.String(d.Id())
	isChanged := false
	if d.HasChange("name") && !d.IsNewResource() {
		isChanged = true
		req.Name = ucloud.String(d.Get("name").(string))
	}
	if d.HasChange("tag") && !d.IsNewResource() {
		isChanged = true
		if value, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(value.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}
	}
	if isChanged {
		if _, err := conn.UpdateSubnetAttribute(req); err != nil {
			return fmt.Errorf("error on %s to subnet %q, %s", "UpdateSubnetAttribute", d.Id(), err)
		}
		d.SetPartial("name")
		d.SetPartial("tag")
		if _, err := subnetWaitForState(client, d.Id()).WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for %s complete to subnet %q, %s", "UpdateSubnetAttribute", d.Id(), err)
		}
	}
	d.Partial(false)
	return resourceUCloudSubnetRead(d, meta)
}

func resourceUCloudSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading subnet, %s", err)
	}
	subnetSet, err := client.describeSubnetById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading subnet %q, %s", d.Id(), err)
	}
	_ = d.Set("name", subnetSet.SubnetName)
	_ = d.Set("cidr_block", fmt.Sprintf("%s/%s", subnetSet.Subnet, subnetSet.Netmask))
	_ = d.Set("vpc_id", subnetSet.VPCId)
	_ = d.Set("tag", subnetSet.Tag)
	_ = d.Set("remark", subnetSet.Remark)
	_ = d.Set("create_time", timestampToString(subnetSet.CreateTime))
	return nil
}

func resourceUCloudSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting subnet, %s", err)
	}
	conn := client.vpcconn
	deleteReq := conn.NewDeleteSubnetRequest()
	deleteReq.SubnetId = ucloud.String(d.Id())
	describeReq := conn.NewDescribeSubnetResourceRequest()
	describeReq.SubnetId = ucloud.String(d.Id())
	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		resp, err := conn.DescribeSubnetResource(describeReq)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on %s before deleting subnet %q, %s", "DescribeSubnetResource", d.Id(), err))
		}
		if len(resp.DataSet) > 0 {
			resourceIDs := []string{}
			for _, item := range resp.DataSet {
				resourceIDs = append(resourceIDs, item.ResourceId)
			}
			return resource.NonRetryableError(fmt.Errorf("error on deleting subnet %q, we find the resource %v bind to it", d.Id(), resourceIDs))
		}
		if _, err := conn.DeleteSubnet(deleteReq); err != nil {
			if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 4411 {
				return resource.RetryableError(fmt.Errorf("error on deleting subnet %q, %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("error on deleting subnet %q, %s", d.Id(), err))
		}
		if _, err := client.describeSubnetById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading subnet when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified subnet %q has not been deleted due to unknown error", d.Id()))
	})
}

func subnetWaitForState(client *productClient, subnetID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			subnetSet, err := client.describeSubnetById(subnetID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return subnetSet, statusInitialized, nil
		},
	}
}
