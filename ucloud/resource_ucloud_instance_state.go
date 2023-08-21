package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"time"
)

func resourceUCloudInstanceState() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudInstanceStateCreate,
		Read:   resourceUCloudInstanceStateRead,
		Update: resourceUCloudInstanceStateUpdate,
		Delete: resourceUCloudInstanceStateDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"stopped",
					"running",
				}, false),
			},
			"force": {
				Type:     schema.TypeBool,
				Required: false,
			},
		},
	}
}

func resourceUCloudInstanceStateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instanceId := d.Get("instance_id").(string)
	state := d.Get("state").(string)
	force := d.Get("force").(bool)

	instance, instanceErr := waitInstanceReady(client, instanceId, d.Timeout(schema.TimeoutCreate))
	if instanceErr != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", instanceErr)
	}

	err := updateInstanceState(client, *instance, state, force)
	if err != nil {
		return err
	}
	_, instanceErr = waitInstanceReady(client, instanceId, d.Timeout(schema.TimeoutCreate))
	if instanceErr != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", instanceErr)
	}
	d.SetId(instanceId)
	return resourceUCloudInstanceStateRead(d, meta)
}

func resourceUCloudInstanceStateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instanceId := d.Id()

	state, err := client.getInstanceState(instanceId)
	if err != nil {
		return err
	}
	d.Set("instance_id", instanceId)
	d.Set("state", state)
	d.Set("force", d.Get("force").(bool))
	return nil
}

func resourceUCloudInstanceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instanceId := d.Id()
	state := d.Get("state").(string)
	force := d.Get("force").(bool)

	instance, instanceErr := waitInstanceReady(client, instanceId, d.Timeout(schema.TimeoutCreate))
	if instanceErr != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", instanceErr)
	}
	err := updateInstanceState(client, *instance, state, force)
	if err != nil {
		return err
	}
	_, instanceErr = waitInstanceReady(client, instanceId, d.Timeout(schema.TimeoutCreate))
	if instanceErr != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", instanceErr)
	}
	return resourceUCloudInstanceStateRead(d, meta)
}

func resourceUCloudInstanceStateDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func waitInstanceReady(client *UCloudClient, id string, timeout time.Duration) (*uhost.UHostInstanceSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending, instanceStatusInitializing, instanceStatusStarting, instanceStatusStopping, instanceStatusRebooting},
		Target:     []string{instanceStatusRunning, instanceStatusStopped},
		Refresh:    getInstanceStateRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*uhost.UHostInstanceSet); ok {
		return output, err
	}

	return nil, err
}

func getInstanceStateRefreshFunc(client *UCloudClient, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := client.describeInstanceById(instanceId)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		state := instance.State
		if state == instanceStatusResizeFail {
			return nil, "", fmt.Errorf("resizing instance failed")
		}

		if state == instanceStatusInstallFail {
			return nil, "", fmt.Errorf("install failed")
		}

		return instance, state, nil
	}
}

func updateInstanceState(client *UCloudClient, instance uhost.UHostInstanceSet, state string, force bool) error {
	switch instance.State {
	case instanceStatusStopped:
		if state == instanceStatusRunning {
			return client.startInstanceById(instance.UHostId)
		}
	case instanceStatusRunning:
		if state == instanceStatusStopped {
			if force {
				return client.poweroffInstanceById(instance.UHostId)
			} else {
				return client.stopInstanceById(instance.UHostId)
			}
		}
	}
	return nil
}
