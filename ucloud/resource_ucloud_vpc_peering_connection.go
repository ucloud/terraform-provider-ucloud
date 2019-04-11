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
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"peer_vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"peer_project_id": {
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
	if v, ok := d.GetOk("peer_project_id"); ok {
		peerProjectId = v.(string)
	}

	req := conn.NewCreateVPCIntercomRequest()
	req.VPCId = ucloud.String(vpcId)
	req.DstVPCId = ucloud.String(peerVpcId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectId)

	_, err := conn.CreateVPCIntercom(req)
	if err != nil {
		return fmt.Errorf("error on creating vpc peering connection, %s", err)
	}

	assocId := fmt.Sprintf(
		"%s@%s#%s:%s@%s#%s",
		client.region, client.projectId, vpcId,
		peerRegion, peerProjectId, peerVpcId,
	)
	d.SetId(assocId)

	// after create vpc peering connection, we need to wait it initialized
	stateConf := vpcConnWaitForState(client, vpcId, peerVpcId, peerRegion, peerProjectId)

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for vpc peering connection %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudVPCPeeringConnectionRead(d, meta)
}

func resourceUCloudVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}

	peerRegion, peerProjectId, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}

	vpcPCSet, err := client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectId)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpc peering connection %q, %s", d.Id(), err)
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
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}

	peerRegion, peerProjectId, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}

	req := conn.NewDeleteVPCIntercomRequest()
	req.VPCId = ucloud.String(assoc.PrimaryId)
	req.DstVPCId = ucloud.String(assoc.ResourceId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectId)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPCIntercom(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpc peering connection %q, %s", d.Id(), err))
		}

		_, err = client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectId)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpc peering connection when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified vpc peering connection %q has not been deleted due to unknown error", d.Id()))
	})
}

func parseVPCPeerDstType(dstType string) (string, string, error) {
	splited := strings.Split(dstType, "@")

	if len(splited) < 2 {
		return "", "", fmt.Errorf(`excepted "region@project_id", got %q`, dstType)
	}

	return splited[0], splited[1], nil
}

func vpcConnWaitForState(client *UCloudClient, vpcId, peerVpcId, peerRegion, peerProjectId string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			v, err := client.describeVPCIntercomById(vpcId, peerVpcId, peerRegion, peerProjectId)
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
