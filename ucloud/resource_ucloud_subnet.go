package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
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
			"cidr_block": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateUCloudCidrBlock,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewCreateSubnetRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	cidrBlock := d.Get("cidr_block").(string)

	// skip parse error, because has been validated at schema validator
	cidr, _ := parseCidrBlock(cidrBlock)
	req.Subnet = ucloud.String(cidr.Network)
	req.Netmask = ucloud.Int(cidr.Mask)

	if val, ok := d.GetOk("name"); ok {
		req.SubnetName = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(val.(string))
	}

	resp, err := conn.CreateSubnet(req)
	if err != nil {
		return fmt.Errorf("error in create subnet, %s", err)
	}

	d.SetId(resp.SubnetId)

	// after create subnet, we need to wait it initialized
	stateConf := subnetWaitForState(client, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("wait for subnet initialize failed in create subnet %s, %s", d.Id(), err)
	}

	return resourceUCloudSubnetUpdate(d, meta)
}

func resourceUCloudSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	d.Partial(true)

	isChanged := false
	req := conn.NewUpdateSubnetAttributeRequest()
	req.SubnetId = ucloud.String(d.Id())

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

	if isChanged {
		_, err := conn.UpdateSubnetAttribute(req)

		if err != nil {
			return fmt.Errorf("do %s failed in update subnet %s, %s", "UpdateSubnetAttribute", d.Id(), err)
		}

		// after update subnet attribute, we need to wait it completed
		stateConf := subnetWaitForState(client, d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("wait for update subnet attribute failed in update subnet %s, %s", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceUCloudSubnetRead(d, meta)
}

func resourceUCloudSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	subnetSet, err := client.describeSubnetById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read subnet %s, %s", "DescribeSubnet", d.Id(), err)
	}

	d.Set("name", subnetSet.SubnetName)
	d.Set("cidr_block", subnetSet.Subnet+"/"+string(subnetSet.Netmask))
	d.Set("vpc_id", subnetSet.VPCId)
	d.Set("tag", subnetSet.Tag)
	d.Set("remark", subnetSet.Remark)
	d.Set("create_time", timestampToString(subnetSet.CreateTime))

	return nil
}

func resourceUCloudSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteSubnetRequest()
	req.SubnetId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteSubnet(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete subnet %s, %s", d.Id(), err))
		}

		_, err := client.describeSubnetById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete subnet %s, %s", "DescribeSubnet", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete subnet but it still exists"))
	})
}

func subnetWaitForState(client *UCloudClient, subnetId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"initialized"},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			subnetSet, err := client.describeSubnetById(subnetId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return subnetSet, "initialized", nil
		},
	}
}
