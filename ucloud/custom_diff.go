package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func diffValidateDefaultSecurityGroup(old, new, meta interface{}) error {
	client := meta.(*UCloudClient)

	// check default firewall is exists when no firewall is specified
	if len(new.(string)) == 0 {
		return client.checkDefaultFirewall()
	}
	return nil
}

func diffValidateUDPNPeerRegion(old, new, meta interface{}) error {
	client := meta.(*UCloudClient)

	if new.(string) == client.region {
		return fmt.Errorf(
			"expected the peering region %s to be different with provider's region %s",
			new.(string), client.region,
		)
	}

	return nil
}

func diffSupressVPCNetworkUpdate(old, new, meta interface{}) error {
	_ = meta.(*UCloudClient)

	o, n := old.(*schema.Set), new.(*schema.Set)
	if o.Difference(n).Len() > 0 && n.Difference(o).Len() > 0 {
		return fmt.Errorf("excepted only create or delete operation for network, could not apply both them, please apply delete first, and then apply create")
	}

	return nil
}
