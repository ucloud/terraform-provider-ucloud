package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"log"
	"testing"
)

func TestAccUCloudIsolationGroup_basic(t *testing.T) {
	var igSet uhost.IsolationGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_isolation_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIsolationGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIsolationGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIsolationGroupExists("ucloud_isolation_group.foo", &igSet),
					testAccCheckIsolationGroupAttributes(&igSet),
					resource.TestCheckResourceAttr("ucloud_isolation_group.foo", "name", "tf-acc-isolation-group-basic"),
					resource.TestCheckResourceAttr("ucloud_isolation_group.foo", "remark", "test"),
				),
			},
		},
	})
}

func testAccCheckIsolationGroupExists(n string, igSet *uhost.IsolationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("isolation group id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeIsolationGroupById(rs.Primary.ID)

		log.Printf("[INFO] isolation group id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*igSet = *ptr
		return nil
	}
}

func testAccCheckIsolationGroupAttributes(igSet *uhost.IsolationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if igSet.GroupId == "" {
			return fmt.Errorf("isolation group id is empty")
		}
		return nil
	}
}

func testAccCheckIsolationGroupDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_isolation_group" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeIsolationGroupById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.GroupId != "" {
			return fmt.Errorf("isolation group still exist")
		}
	}

	return nil
}

const testAccIsolationGroupConfig = `
resource "ucloud_isolation_group" "foo" {
	name  = "tf-acc-isolation-group-basic"
	remark = "test"
}
`
