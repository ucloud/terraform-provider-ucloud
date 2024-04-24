package ucloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func testAccCheckLabelDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_label" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		key, value, err := parseUCloudLabelID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("fail to parse id: %v", rs.Primary.ID)
		}
		_, err = client.describeLabel(key, value)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("label still exist")
	}

	return nil
}

const testAccLabelConfig = `
resource "ucloud_label" "foo" {
	key  = "tf-acc-label-key"
	value  = "tf-acc-label-value"
}
`
