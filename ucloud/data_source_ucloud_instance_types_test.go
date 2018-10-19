package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudInstanceTypesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataInstanceTypesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_instance_types.foo"),
					resource.TestCheckResourceAttr("data.ucloud_instance_types.foo", "instance_types.0.id", "n-basic-2"),
					resource.TestCheckResourceAttr("data.ucloud_instance_types.foo", "instance_types.1.id", "n-customize-2-4"),
				),
			},
		},
	})
}

const testAccDataInstanceTypesConfig = `
data "ucloud_instance_types" "foo" {
	cpu = 2
	memory = 4
}
`
