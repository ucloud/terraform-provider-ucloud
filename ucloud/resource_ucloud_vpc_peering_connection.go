package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudVPCPeeringConnectionCreate,
		Read:   resourceUCloudVPCPeeringConnectionRead,
		Delete: resourceUCloudVPCPeeringConnectionDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"peer_vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"peer_project_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUCloudVPCPeeringConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	vpcId := d.Get("vpc_id").(string)
	peerVpcId := d.Get("peer_vpc_id").(string)
	peerRegion := client.region

	peerProjectId := client.projectId
	if val, ok := d.GetOk("peer_project_id"); ok {
		peerProjectId = val.(string)
	}

	req := conn.NewCreateVPCIntercomRequest()
	req.VPCId = ucloud.String(vpcId)
	req.DstVPCId = ucloud.String(peerVpcId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectId)

	_, err := conn.CreateVPCIntercom(req)
	if err != nil {
		return fmt.Errorf("error in create vpc peering connection, %s", err)
	}

	assocId := fmt.Sprintf(
		"%s@%s#%s:%s@%s#%s",
		client.region, client.projectId, vpcId,
		peerRegion, peerProjectId, peerVpcId,
	)
	d.SetId(assocId)

	// after create vpc peering connection, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"initialized"},
		Timeout:    5 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			vpcPCSet, err := client.describeVPCIntercomById(vpcId, peerVpcId, peerRegion, peerProjectId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return vpcPCSet, "initialized", nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("wait for vpc peering connection initialize failed in create vpc peering connection %s, %s", d.Id(), err)
	}

	return resourceUCloudVPCPeeringConnectionRead(d, meta)
}

func resourceUCloudVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error in parse vpc peering connection %s, %s", d.Id(), err)
	}

	peerRegion, peerProjectId, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error in parse vpc peering connection %s, %s", d.Id(), err)
	}

	vpcPCSet, err := client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectId)

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("do %s failed in read vpc peering connection %s, %s", "DescribeVPCIntercom", d.Id(), err)
	}

	d.Set("vpc_id", d.Get("vpc_id").(string))
	d.Set("peer_vpc_id", vpcPCSet.VPCId)
	d.Set("peer_project_id", vpcPCSet.ProjectId)

	return nil
}

func resourceUCloudVPCPeeringConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.vpcconn

	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error in parse vpc peering connection %s, %s", d.Id(), err)
	}
	peerRegion, peerProjectId, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error in parse vpc peering connection %s, %s", d.Id(), err)
	}

	req := conn.NewDeleteVPCIntercomRequest()
	req.VPCId = ucloud.String(assoc.PrimaryId)
	req.DstVPCId = ucloud.String(assoc.ResourceId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectId)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		// retry by sdk implementations
		if _, err := conn.DeleteVPCIntercom(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error in delete vpc peering connection %s, %s", d.Id(), err))
		}

		_, err = client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectId)

		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("do %s failed in delete vpc peering connection %s, %s", "DescribeVPCIntercom", d.Id(), err))
		}

		// delete but it still exists
		return resource.RetryableError(fmt.Errorf("delete vpc peering connection but it still exists"))
	})
}

func parseVPCPeerDstType(dstType string) (string, string, error) {
	splited := strings.Split(dstType, "@")

	if len(splited) < 2 {
		return "", "", fmt.Errorf(`excepted "region@project_id", got %s`, dstType)
	}

	return splited[0], splited[1], nil
}
