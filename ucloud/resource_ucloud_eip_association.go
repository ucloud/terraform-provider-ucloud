package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudEIPAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudEIPAssociationCreate,
		Read:   resourceUCloudEIPAssociationRead,
		Delete: resourceUCloudEIPAssociationDelete,

		SchemaVersion: 1,
		MigrateState:  resourceUCloudEIPAssociationMigrateState,

		Schema: map[string]*schema.Schema{
			"eip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if (isStringIn(old, []string{resourceTypeInstance, eipResourceTypeUHost}) && isStringIn(new, []string{resourceTypeInstance, eipResourceTypeUHost})) ||
						(isStringIn(old, []string{resourceTypeLb, eipResourceTypeULB}) && isStringIn(new, []string{resourceTypeLb, eipResourceTypeULB})) ||
						(isStringIn(old, []string{resourceTypeBareMetal, eipResourceTypeUPHost}) && isStringIn(new, []string{resourceTypeBareMetal, eipResourceTypeUPHost})) {
						return true
					}
					return false
				},
			},

			"resource_id": {
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
	resourceId := d.Get("resource_id").(string)
	var resourceType string
	if len(strings.Split(resourceId, "-")) > 0 {
		resourceType = strings.Split(resourceId, "-")[0]
	}

	if v, ok := d.GetOk("resource_type"); ok {
		resourceType = lowerCaseProdCvt.convert(v.(string))
	}

	if len(resourceType) == 0 {
		return fmt.Errorf("must set `resource_type` when creating eip association")
	}

	req := conn.NewBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceType = ucloud.String(resourceType)
	req.ResourceId = ucloud.String(resourceId)

	_, err := conn.BindEIP(req)
	if err != nil {
		return fmt.Errorf("error on creating eip association, %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", eipId, resourceId))

	// after bind eip we need to wait it completed
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{eipStatusUsed},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			eip, err := client.describeEIPById(eipId)
			if err != nil {
				return nil, "", err
			}

			state := eip.Status
			if state != eipStatusUsed {
				state = statusPending
			}

			return eip, state, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for eip association is completed when creating %q, %s", d.Id(), err)
	}

	return resourceUCloudEIPAssociationRead(d, meta)
}

func resourceUCloudEIPAssociationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	p := strings.Split(d.Id(), ":")
	res, err := client.describeEIPResourceById(p[0], p[1])
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading eip association %q, %s", d.Id(), err)
	}

	// remote api has not returned eip
	d.Set("eip_id", d.Get("eip_id"))
	d.Set("resource_id", res.ResourceID)
	if v, ok := d.GetOk("resource_type"); ok && isStringIn(v.(string), []string{resourceTypeInstance, resourceTypeLb}) {
		d.Set("resource_type", lowerCaseProdCvt.unconvert(res.ResourceType))
	} else {
		d.Set("resource_type", res.ResourceType)
	}

	return nil
}

func resourceUCloudEIPAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	p := strings.Split(d.Id(), ":")
	req := conn.NewUnBindEIPRequest()
	req.EIPId = ucloud.String(p[0])
	req.ResourceId = ucloud.String(p[1])
	resourceType := eipResourceTypeUHost
	if len(strings.Split(p[1], "-")) > 0 {
		resourceType = strings.Split(p[1], "-")[0]
	}

	if v, ok := d.GetOk("resource_type"); ok {
		resourceType = lowerCaseProdCvt.convert(v.(string))
	}
	req.ResourceType = ucloud.String(resourceType)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.describeEIPResourceById(p[0], p[1])
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading eip association before deleting %q, %s", d.Id(), err))
		}

		if _, err := conn.UnBindEIP(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting eip association %q, %s", d.Id(), err))
		}

		_, err = client.describeEIPResourceById(p[0], p[1])
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error on reading eip association when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified eip association %q has not been deleted due to unknown error", d.Id()))
	})
}
