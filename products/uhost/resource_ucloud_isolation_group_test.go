package uhost_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
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

const testAccIsolationGroupConfig = `
resource "ucloud_isolation_group" "foo" {
	name  = "tf-acc-isolation-group-basic"
	remark = "test"
}
`
