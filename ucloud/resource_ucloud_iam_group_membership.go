package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceUCloudIAMGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudIAMGroupMembershipCreate,
		Update: resourceUCloudIAMGroupMembershipUpdate,
		Read:   resourceUCloudIAMGroupMembershipRead,
		Delete: resourceUCloudIAMGroupMembershipDelete,

		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_names": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceUCloudIAMGroupMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	group := d.Get("group_name").(string)
	users := interfaceSliceToStringSlice(d.Get("user_names").(*schema.Set).List())

	err := client.addUsersToGroup(users, group)
	if err != nil {
		return fmt.Errorf("error on add users to group, %s", err)
	}
	d.SetId(group)
	return resourceUCloudIAMGroupMembershipRead(d, meta)
}

func resourceUCloudIAMGroupMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	if d.HasChange("user_names") {
		d.SetPartial("user_names")
		o, n := d.GetChange("user_names")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		oldSet := o.(*schema.Set)
		newSet := n.(*schema.Set)

		remove := interfaceSliceToStringSlice(oldSet.Difference(newSet).List())
		add := interfaceSliceToStringSlice(newSet.Difference(oldSet).List())
		group := d.Id()

		if err := client.removeUsersFromGroup(remove, group); err != nil {
			return fmt.Errorf("error on update membership when remove users from group, %s", err)
		}

		if err := client.addUsersToGroup(add, group); err != nil {
			return fmt.Errorf("error on update membership when remove users from group, %s", err)
		}
	}

	d.Partial(false)
	return resourceUCloudIAMGroupMembershipRead(d, meta)
}

func resourceUCloudIAMGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	resp, err := client.describeGroupMembership(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading group membership %q, %s", d.Id(), err)
	}
	var users []string
	for _, v := range resp {
		users = append(users, v.UserName)
	}
	d.Set("group_name", d.Id())
	d.Set("user_names", users)
	return nil
}

func resourceUCloudIAMGroupMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	group := d.Get("group_name").(string)
	users := interfaceSliceToStringSlice(d.Get("user_names").(*schema.Set).List())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if err := client.removeUsersFromGroup(users, group); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on remove users from group %q, %s", d.Id(), err))
		}

		_, err := client.describeGroupMembership(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading group membership when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified group membership %q has not been deleted due to unknown error", d.Id()))
	})
}
