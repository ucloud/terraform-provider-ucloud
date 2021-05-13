package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"log"
	"testing"
)

func TestAccUCloudUS3Bucket_basic(t *testing.T) {
	var us3BucketSet ufile.UFileBucketSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_us3_bucket.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUS3BucketDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUS3BucketConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUS3BucketExists("ucloud_us3_bucket.foo", &us3BucketSet),
					testAccCheckUS3BucketAttributes(&us3BucketSet),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "name", "tf-acc-us3-bucket-basic"),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "type", "private"),
				),
			},
			{
				Config: testAccUS3BucketConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUS3BucketExists("ucloud_us3_bucket.foo", &us3BucketSet),
					testAccCheckUS3BucketAttributes(&us3BucketSet),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "name", "tf-acc-us3-bucket-basic"),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "type", "public"),
				),
			},
		},
	})
}

func testAccCheckUS3BucketExists(n string, us3BucketSet *ufile.UFileBucketSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("us3 bucket id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUS3BucketById(rs.Primary.ID)

		log.Printf("[INFO] us3 bucket id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*us3BucketSet = *ptr
		return nil
	}
}

func testAccCheckUS3BucketAttributes(us3BucketSet *ufile.UFileBucketSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if us3BucketSet.BucketName == "" {
			return fmt.Errorf("us3 bucket id is empty")
		}
		return nil
	}
}

func testAccCheckUS3BucketDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_us3_bucket" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUS3BucketById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.BucketName != "" {
			return fmt.Errorf("us3 bucket still exist")
		}
	}

	return nil
}

const testAccUS3BucketConfig = `
resource "ucloud_us3_bucket" "foo" {
	name  	 	  = "tf-acc-us3-bucket-basic"
	type      	  = "private"
}
`

const testAccUS3BucketConfigUpdate = `
resource "ucloud_us3_bucket" "foo" {
	name  	 	  = "tf-acc-us3-bucket-basic"
	type      	  = "public"
}
`
