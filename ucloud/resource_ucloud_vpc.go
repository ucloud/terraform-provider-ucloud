package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudVPC() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPCCreate,
		Read:   resourceUCloudVPCRead,
		Update: resourceUCloudVPCUpdate,
		Delete: resourceUCloudVPCDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"cidr_blocks": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateUCloudCidrBlock,
				},
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

			"network_info": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"update_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudVPCCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).vpcconn

	req := conn.NewCreateVPCRequest()
	req.Name = ucloud.String(d.Get("name").(string))
	req.Network = ifaceToStringSlice(d.Get("cidr_blocks"))

	if val, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(val.(string))
	}

	resp, err := conn.CreateVPC(req)

	if err != nil {
		return fmt.Errorf("error in create vpc, %s", err)
	}

	d.SetId(resp.VPCId)

	time.Sleep(2 * time.Second)

	return resourceUCloudVPCUpdate(d, meta)
}

func resourceUCloudVPCUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO:need backend API support
	return resourceUCloudVPCRead(d, meta)
}

func resourceUCloudVPCRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	ins, err := client.describeVPCById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read vpc %s, %s", "DescribeVPC", d.Id(), err)
	}

	d.Set("name", ins.Name)
	d.Set("tag", ins.Tag)

	// TODO: [API-ERROR] remark is not in api model, should be checked!
	// d.Set("remark", ins.Remark)

	d.Set("cidr_blocks", ins.Network)
	d.Set("create_time", timestampToString(ins.CreateTime))
	d.Set("update_time", timestampToString(ins.UpdateTime))

	networkInfo := []map[string]interface{}{}
	for _, item := range ins.NetworkInfo {
		networkInfo = append(networkInfo, map[string]interface{}{
			"cidr_block": item.Network,
		})
	}
	d.Set("network_info", networkInfo)

	return nil
}

func resourceUCloudVPCDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteVPCRequest()
	req.VPCId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPC(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete vpc %s, %s", d.Id(), err))
		}

		_, err := client.describeVPCById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete vpc %s, %s", "DescribeVPC", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete vpc but it still exists"))
	})
}
