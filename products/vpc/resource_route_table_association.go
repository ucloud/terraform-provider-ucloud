package vpc

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudRouteTableAssociationCreate,
		Update: resourceUCloudRouteTableAssociationUpdate,
		Read:   resourceUCloudRouteTableAssociationRead,
		Delete: resourceUCloudRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceUCloudRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating route table association, %s", err)
	}
	conn := client.vpcconn

	subnetID := d.Get("subnet_id").(string)
	routeTableID := d.Get("route_table_id").(string)

	req := conn.NewAssociateRouteTableRequest()
	req.SubnetId = ucloud.String(subnetID)
	req.RouteTableId = ucloud.String(routeTableID)
	if _, err := conn.AssociateRouteTable(req); err != nil {
		return fmt.Errorf("error on associating subnet %q with route table %q, %s", subnetID, routeTableID, err)
	}

	d.SetId(subnetID)
	return resourceUCloudRouteTableAssociationRead(d, meta)
}

func resourceUCloudRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating route table association, %s", err)
	}
	conn := client.vpcconn

	if d.HasChange("route_table_id") && !d.IsNewResource() {
		subnetID := d.Get("subnet_id").(string)
		routeTableID := d.Get("route_table_id").(string)
		req := conn.NewAssociateRouteTableRequest()
		req.SubnetId = ucloud.String(subnetID)
		req.RouteTableId = ucloud.String(routeTableID)
		if _, err := conn.AssociateRouteTable(req); err != nil {
			return fmt.Errorf("error on associating subnet %q with route table %q, %s", subnetID, routeTableID, err)
		}
		d.SetPartial("route_table_id")
	}

	return resourceUCloudRouteTableAssociationRead(d, meta)
}

func resourceUCloudRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading route table association, %s", err)
	}
	subnet, err := client.describeSubnetById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading route table association %q, %s", d.Id(), err)
	}
	_ = d.Set("subnet_id", subnet.SubnetId)
	_ = d.Set("route_table_id", subnet.RouteTableId)
	return nil
}

func resourceUCloudRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting route table association, %s", err)
	}
	conn := client.vpcconn

	subnetID := d.Get("subnet_id").(string)
	req := conn.NewAssociateRouteTableRequest()
	req.SubnetId = ucloud.String(subnetID)
	// Leaving RouteTableId empty rebinds the subnet to its default route table,
	// which is the "disassociated" state for a custom route table binding.
	if _, err := conn.AssociateRouteTable(req); err != nil {
		return fmt.Errorf("error on disassociating subnet %q from its route table, %s", subnetID, err)
	}
	return nil
}
