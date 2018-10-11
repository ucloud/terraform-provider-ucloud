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
	conn := meta.(*UCloudClient).vpcconn

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

	time.Sleep(2 * time.Second)

	return resourceUCloudSubnetUpdate(d, meta)
}

func resourceUCloudSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).vpcconn

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

		time.Sleep(2 * time.Second)
	}

	d.Partial(false)

	return resourceUCloudSubnetRead(d, meta)
}

func resourceUCloudSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	ins, err := client.describeSubnetById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read subnet %s, %s", "DescribeSubnet", d.Id(), err)
	}

	d.Set("name", ins.SubnetName)
	d.Set("cidr_block", ins.Subnet+"/"+string(ins.Netmask))
	d.Set("vpc_id", ins.VPCId)
	d.Set("tag", ins.Tag)
	d.Set("remark", ins.Remark)
	d.Set("create_time", timestampToString(ins.CreateTime))

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
