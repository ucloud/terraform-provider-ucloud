package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudRouteTableAssociation_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationConfig,
			},
			{
				ResourceName:      "ucloud_route_table_association.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
