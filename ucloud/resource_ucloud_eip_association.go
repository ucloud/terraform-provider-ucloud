package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudEIPAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudEIPAssociationCreate,
		Read:   resourceUCloudEIPAssociationRead,
		Delete: resourceUCloudEIPAssociationDelete,

		Schema: map[string]*schema.Schema{
			"eip_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUCloudEIPAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	eipId := d.Get("eip_id").(string)
	resourceType := ulbMap.convert(uhostMap.convert(d.Get("resource_type").(string)))
	resourceId := d.Get("resource_id").(string)

	req := conn.NewBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceType = ucloud.String(resourceType)
	req.ResourceId = ucloud.String(resourceId)

	_, err := conn.BindEIP(req)
	if err != nil {
		return fmt.Errorf("error in create eip association, %s", err)
	}

	d.SetId(fmt.Sprintf("eip#%s:%s#%s", eipId, resourceType, resourceId))

	// after bind eip we need to wait it completed
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"used"},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			eip, err := client.describeEIPById(eipId)
			if err != nil {
				return nil, "", err
			}

			state := eip.Status
			if state != "used" {
				state = "pending"
			}

			return eip, state, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("wait for bind eip failed in create eip association %s, %s", d.Id(), err)
	}

	return resourceUCloudEIPAssociationRead(d, meta)
}

func resourceUCloudEIPAssociationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error in parse eip association %s, %s", d.Id(), err)
	}

	resource, err := client.describeEIPResourceById(assoc.PrimaryId, assoc.ResourceType, assoc.ResourceId)

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read eip association %s, %s", "DescribeEIP", d.Id(), err)
	}
	//TODO:[API-ERROR] UnetEIPResourceSet don't have EIPId
	d.Set("eip_id", d.Get("eip_id"))
	d.Set("resource_id", resource.ResourceId)
	d.Set("resource_type", ulbMap.unconvert(uhostMap.unconvert(resource.ResourceType)))

	return nil
}

func resourceUCloudEIPAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error in parse eip association %s, %s", d.Id(), err)
	}

	req := conn.NewUnBindEIPRequest()
	req.EIPId = ucloud.String(assoc.PrimaryId)
	req.ResourceId = ucloud.String(assoc.ResourceId)
	req.ResourceType = ucloud.String(assoc.ResourceType)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.UnBindEIP(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete eip association %s, %s", d.Id(), err))
		}

		_, err := client.describeEIPResourceById(assoc.PrimaryId, assoc.ResourceType, assoc.ResourceId)

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete eip association %s, %s", "DescribeEIP", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("delete eip association but it still exists"))
	})
}
