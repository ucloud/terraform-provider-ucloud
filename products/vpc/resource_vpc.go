package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudVPC() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPCCreate,
		Update: resourceUCloudVPCUpdate,
		Read:   resourceUCloudVPCRead,
		Delete: resourceUCloudVPCDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("cidr_blocks", diffSupressVPCNetworkUpdate),
		),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},
			"cidr_blocks": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCIDRBlock,
				},
				Set: hashCIDR,
			},
			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},
			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"cidr_block": {Type: schema.TypeString, Computed: true},
				}},
			},
			"update_time": {Type: schema.TypeString, Computed: true},
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudVPCCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating vpc, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewCreateVPCRequest()
	req.Network = schemaSetToStringSlice(d.Get("cidr_blocks"))
	if value, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(value.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-vpc-"))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}
	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
	}
	resp, err := conn.CreateVPC(req)
	if err != nil {
		return fmt.Errorf("error on creating vpc, %s", err)
	}
	d.SetId(resp.VPCId)
	if _, err = vpcWaitForState(client, d.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for vpc %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudVPCRead(d, meta)
}

func resourceUCloudVPCRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading vpc, %s", err)
	}
	vpcSet, err := client.describeVPCById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpc %q, %s", d.Id(), err)
	}
	_ = d.Set("name", vpcSet.Name)
	_ = d.Set("tag", vpcSet.Tag)
	_ = d.Set("cidr_blocks", vpcSet.Network)
	_ = d.Set("create_time", timestampToString(vpcSet.CreateTime))
	_ = d.Set("update_time", timestampToString(vpcSet.UpdateTime))
	networkInfo := []map[string]interface{}{}
	for _, item := range vpcSet.NetworkInfo {
		networkInfo = append(networkInfo, map[string]interface{}{"cidr_block": item.Network})
	}
	if err := d.Set("network_info", networkInfo); err != nil {
		return err
	}
	return nil
}

func resourceUCloudVPCUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating vpc, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	if d.HasChange("cidr_blocks") && !d.IsNewResource() {
		oldValue, newValue := d.GetChange("cidr_blocks")
		oldSet, newSet := oldValue.(*schema.Set), newValue.(*schema.Set)
		if added := newSet.Difference(oldSet); added.Len() > 0 {
			req := conn.NewAddVPCNetworkRequest()
			req.VPCId = ucloud.String(d.Id())
			req.Network = schemaSetToStringSlice(added)
			if _, err := conn.AddVPCNetwork(req); err != nil {
				return fmt.Errorf("error on %s to vpc %q, %s", "AddVPCNetwork", d.Id(), err)
			}
		}
		if removed := oldSet.Difference(newSet); removed.Len() > 0 {
			req := conn.NewUpdateVPCNetworkRequest()
			req.VPCId = ucloud.String(d.Id())
			req.Network = schemaSetToStringSlice(newSet)
			if _, err := conn.UpdateVPCNetwork(req); err != nil {
				return fmt.Errorf("error on %s to vpc %q, %s", "UpdateVPCNetwork", d.Id(), err)
			}
		}
		d.SetPartial("cidr_blocks")
	}
	d.Partial(false)
	return resourceUCloudVPCRead(d, meta)
}

func resourceUCloudVPCDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting vpc, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewDeleteVPCRequest()
	req.VPCId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPC(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpc %q, %s", d.Id(), err))
		}
		if _, err := client.describeVPCById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpc when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified vpc %q has not been deleted due to unknown error", d.Id()))
	})
}

func vpcWaitForState(client *productClient, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			vpcSet, err := client.describeVPCById(id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return vpcSet, statusInitialized, nil
		},
	}
}

func diffSupressVPCNetworkUpdate(old, new, meta interface{}) error {
	_ = meta
	oldSet, newSet := old.(*schema.Set), new.(*schema.Set)
	if oldSet.Difference(newSet).Len() > 0 && newSet.Difference(oldSet).Len() > 0 {
		return fmt.Errorf("excepted only create or delete operation for network, could not apply both them, please apply delete first, and then apply create")
	}
	return nil
}
