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
		Delete: resourceUCloudVPCDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      resource.PrefixedUniqueId("tf-vpc-"),
				ValidateFunc: validateName,
			},

			"cidr_blocks": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateUCloudCidrBlock,
				},
				Set: hashCIDR,
			},

			"tag": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"remark": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewCreateVPCRequest()
	req.Name = ucloud.String(d.Get("name").(string))
	req.Network = schemaSetToStringSlice(d.Get("cidr_blocks"))

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	if v, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(v.(string))
	}

	resp, err := conn.CreateVPC(req)
	if err != nil {
		return fmt.Errorf("error on creating vpc, %s", err)
	}

	d.SetId(resp.VPCId)

	// after create vpc, we need to wait it initialized
	_, err = vpcWaitForState(client, d.Id()).WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for vpc %s complete creating, %s", d.Id(), err)
	}

	return resourceUCloudVPCRead(d, meta)
}

func resourceUCloudVPCRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	vpcSet, err := client.describeVPCById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpc %s, %s", d.Id(), err)
	}

	d.Set("name", vpcSet.Name)
	d.Set("tag", vpcSet.Tag)

	// TODO: [API-ERROR] remark is not in api model, should be checked!
	// d.Set("remark", vpcSet.Remark)

	d.Set("cidr_blocks", vpcSet.Network)
	d.Set("create_time", timestampToString(vpcSet.CreateTime))
	d.Set("update_time", timestampToString(vpcSet.UpdateTime))

	networkInfo := []map[string]interface{}{}
	for _, item := range vpcSet.NetworkInfo {
		networkInfo = append(networkInfo, map[string]interface{}{
			"cidr_block": item.Network,
		})
	}

	if err := d.Set("network_info", networkInfo); err != nil {
		return err
	}

	return nil
}

func resourceUCloudVPCDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	req := conn.NewDeleteVPCRequest()
	req.VPCId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPC(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpc %s, %s", d.Id(), err))
		}

		_, err := client.describeVPCById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpc when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vpc %s has not been deleted due to unknown error", d.Id()))
	})
}

func vpcWaitForState(client *UCloudClient, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			v, err := client.describeVPCById(id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			return v, statusInitialized, nil
		},
	}
}
