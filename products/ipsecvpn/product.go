package ipsecvpn

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "ipsecvpn"

type adapterV1 struct{}

var _ product.V1 = adapterV1{}

// New returns the stable product registration consumed by the provider.
func New() product.V1 {
	return adapterV1{}
}

func (adapterV1) Registration() product.Registration {
	return product.Registration{
		Name: Name,
		Resources: map[string]*schema.Resource{
			"ucloud_vpn_gateway":          resourceUCloudVPNGateway(),
			"ucloud_vpn_customer_gateway": resourceUCloudVPNCustomerGateway(),
			"ucloud_vpn_connection":       resourceUCloudVPNConnection(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_vpn_gateways":          dateSourceUCloudVPNGateways(),
			"ucloud_vpn_customer_gateways": dateSourceUCloudVPNCustomerGateways(),
			"ucloud_vpn_connections":       dateSourceUCloudVPNConnections(),
		},
	}
}
