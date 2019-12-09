package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLBSSL_basic(t *testing.T) {
	var sslSet ulb.ULBSSLSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb_ssl.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBSSLDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLBSSLConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLExists("ucloud_lb_ssl.foo", &sslSet),
					testAccCheckLBSSLAttributes(&sslSet),
					resource.TestCheckResourceAttr("ucloud_lb_ssl.foo", "name", "tf-acc-lb-ssl"),
				),
			},
		},
	})

}

func testAccCheckLBSSLExists(n string, sslSet *ulb.ULBSSLSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("lb ssl id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeLBSSLById(rs.Primary.ID)

		log.Printf("[INFO] lb ssl id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*sslSet = *ptr
		return nil
	}
}

func testAccCheckLBSSLAttributes(sslSet *ulb.ULBSSLSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sslSet.SSLId == "" {
			return fmt.Errorf("lb ssl id is empty")
		}
		return nil
	}
}

func testAccCheckLBSSLDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb_ssl" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeLBSSLById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SSLId != "" {
			return fmt.Errorf("lb ssl still exist")
		}
	}

	return nil
}

const testAccLBSSLConfig = `
resource "ucloud_lb_ssl" "foo" {
	name 		= "tf-acc-lb-ssl"
	private_key = "${file("test-fixtures/private.key")}"
	user_cert 	= "${file("test-fixtures/user.crt")}"
	ca_cert 	= "${file("test-fixtures/ca.crt")}"
}
`
