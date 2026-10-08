package vpc

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudRouteTableCreate,
		Update: resourceUCloudRouteTableUpdate,
		Read:   resourceUCloudRouteTableRead,
		Delete: resourceUCloudRouteTableDelete,
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
			"vpc_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnet_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"create_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceUCloudRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating route table, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewCreateRouteTableRequest()
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))

	var name string
	if value, ok := d.GetOk("name"); ok {
		name = value.(string)
	} else {
		name = resource.PrefixedUniqueId("tf-route-table-")
	}
	req.Name = ucloud.String(name)

	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}
	if value, ok := d.GetOk("remark"); ok {
		req.Remark = ucloud.String(value.(string))
	}
	resp, err := conn.CreateRouteTable(req)
	if err != nil {
		return fmt.Errorf("error on creating route table, %s", err)
	}
	d.SetId(resp.RouteTableId)
	// The public API does not return the route table name on read, so persist
	// the generated/configured name here to keep Terraform state consistent.
	_ = d.Set("name", name)
	return resourceUCloudRouteTableRead(d, meta)
}

func resourceUCloudRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating route table, %s", err)
	}
	conn := client.vpcconn
	d.Partial(true)
	req := conn.NewUpdateRouteTableAttributeRequest()
	req.RouteTableId = ucloud.String(d.Id())
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
	if d.HasChange("remark") && !d.IsNewResource() {
		isChanged = true
		req.Remark = ucloud.String(d.Get("remark").(string))
	}
	if isChanged {
		if _, err := conn.UpdateRouteTableAttribute(req); err != nil {
			return fmt.Errorf("error on %s to route table %q, %s", "UpdateRouteTableAttribute", d.Id(), err)
		}
		d.SetPartial("name")
		d.SetPartial("tag")
		d.SetPartial("remark")
	}
	d.Partial(false)
	return resourceUCloudRouteTableRead(d, meta)
}

func resourceUCloudRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading route table, %s", err)
	}
	routeTable, err := client.describeRouteTableById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading route table %q, %s", d.Id(), err)
	}
	_ = d.Set("vpc_id", routeTable.VPCId)
	_ = d.Set("vpc_name", routeTable.VPCName)
	_ = d.Set("tag", routeTable.Tag)
	_ = d.Set("remark", routeTable.Remark)
	_ = d.Set("route_table_type", routeTable.RouteTableType)
	_ = d.Set("subnet_count", routeTable.SubnetCount)
	_ = d.Set("subnet_ids", routeTable.SubnetIds)
	_ = d.Set("create_time", timestampToString(routeTable.CreateTime))
	return nil
}

func resourceUCloudRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting route table, %s", err)
	}
	conn := client.vpcconn
	req := conn.NewDeleteRouteTableRequest()
	req.RouteTableId = ucloud.String(d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteRouteTable(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting route table %q, %s", d.Id(), err))
		}
		if _, err := client.describeRouteTableById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading route table when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified route table %q has not been deleted due to unknown error", d.Id()))
	})
}
