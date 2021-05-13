package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudUS3BucketsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUS3BucketsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_us3_buckets.foo"),
					resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.0.name", "tf-acc-us3-buckets-datasource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.0.tag", "Default"),
				),
			},
		},
	})
}

const testAccDataUS3BucketsConfig = `
variable "name" {
	default = "tf-acc-us3-buckets-datasource-basic"
}

resource "ucloud_us3_bucket" "foo" {
	name          = "${var.name}"
	type  		  = "private"
}

data "ucloud_us3_buckets" "foo" {
	name_regex  = "${ucloud_us3_bucket.foo.name}"
}
`
