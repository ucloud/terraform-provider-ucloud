package label_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudLabel_basic(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_label.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLabelDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLabelConfig,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_label.foo", "key", "tf-acc-label-key"),
					resource.TestCheckResourceAttr("ucloud_label.foo", "value", "tf-acc-label-value"),
				),
			},
		},
	})
}

const testAccLabelConfig = `
resource "ucloud_label" "foo" {
	key  = "tf-acc-label-key"
	value  = "tf-acc-label-value"
}
`
