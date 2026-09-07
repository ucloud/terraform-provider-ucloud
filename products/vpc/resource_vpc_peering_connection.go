package vpc

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				ForceNew: true,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},
	}
}

func resourceUCloudVPCPeeringConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating vpc peering connection, %s", err)
	}
	conn := client.vpcconn
	vpcID := d.Get("vpc_id").(string)
	peerVPCID := d.Get("peer_vpc_id").(string)
	peerRegion := client.region
	peerProjectID := client.projectId
	if value, ok := d.GetOk("peer_project_id"); ok {
		peerProjectID = value.(string)
	}
	if value, ok := d.GetOk("peer_region"); ok {
		peerRegion = value.(string)
	}
	req := conn.NewCreateVPCIntercomRequest()
	req.VPCId = ucloud.String(vpcID)
	req.DstVPCId = ucloud.String(peerVPCID)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectID)
	if _, err := conn.CreateVPCIntercom(req); err != nil {
		return fmt.Errorf("error on creating vpc peering connection, %s", err)
	}
	d.SetId(fmt.Sprintf("%s@%s#%s:%s@%s#%s", client.region, client.projectId, vpcID, peerRegion, peerProjectID, peerVPCID))
	if _, err := vpcConnWaitForState(client, vpcID, peerVPCID, peerRegion, peerProjectID).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for vpc peering connection %q complete creating, %s", d.Id(), err)
	}
	return resourceUCloudVPCPeeringConnectionRead(d, meta)
}

func resourceUCloudVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading vpc peering connection, %s", err)
	}
	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}
	peerRegion, peerProjectID, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}
	peer, err := client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectID)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading vpc peering connection %q, %s", d.Id(), err)
	}
	_ = d.Set("vpc_id", d.Get("vpc_id").(string))
	_ = d.Set("peer_vpc_id", peer.VPCId)
	_ = d.Set("peer_project_id", peer.ProjectId)
	_ = d.Set("peer_region", peer.DstRegion)
	return nil
}

func resourceUCloudVPCPeeringConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting vpc peering connection, %s", err)
	}
	conn := client.vpcconn
	assoc, err := parseAssociationInfo(d.Id())
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}
	peerRegion, peerProjectID, err := parseVPCPeerDstType(assoc.ResourceType)
	if err != nil {
		return fmt.Errorf("error on parsing vpc peering connection %q, %s", d.Id(), err)
	}
	req := conn.NewDeleteVPCIntercomRequest()
	req.VPCId = ucloud.String(assoc.PrimaryId)
	req.DstVPCId = ucloud.String(assoc.ResourceId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectID)
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVPCIntercom(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting vpc peering connection %q, %s", d.Id(), err))
		}
		if _, err := client.describeVPCIntercomById(assoc.PrimaryId, assoc.ResourceId, peerRegion, peerProjectID); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading vpc peering connection when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified vpc peering connection %q has not been deleted due to unknown error", d.Id()))
	})
}

func parseVPCPeerDstType(dstType string) (string, string, error) {
	split := strings.Split(dstType, "@")
	if len(split) < 2 {
		return "", "", fmt.Errorf(`excepted "region@project_id", got %q`, dstType)
	}
	return split[0], split[1], nil
}

func vpcConnWaitForState(client *productClient, vpcID, peerVPCID, peerRegion, peerProjectID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			peer, err := client.describeVPCIntercomById(vpcID, peerVPCID, peerRegion, peerProjectID)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			return peer, statusInitialized, nil
		},
	}
}
