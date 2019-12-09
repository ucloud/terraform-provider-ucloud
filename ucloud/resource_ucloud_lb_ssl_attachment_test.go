package ucloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLBSSLAttachment_basic(t *testing.T) {
	var sslSet ulb.ULBSSLSet
	var lbSet ulb.ULBSet
	var vserverSet ulb.ULBVServerSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb_ssl_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBSSLAttachmentDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLBSSLAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckLBSSLExists("ucloud_lb_ssl.foo", &sslSet),
					testAccCheckLBSSLAttachmentExists("ucloud_lb_ssl_attachment.foo", &sslSet, &lbSet, &vserverSet),
				),
			},
		},
	})
}

func testAccCheckLBSSLAttachmentExists(n string, sslSet *ulb.ULBSSLSet, lbSet *ulb.ULBSet, vserverSet *ulb.ULBVServerSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("lb ssl attachment id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)

		return resource.Retry(3*time.Minute, func() *resource.RetryError {
			d, err := client.describeLBSSLAttachmentById(
				rs.Primary.Attributes["ssl_id"],
				rs.Primary.Attributes["load_balancer_id"],
				rs.Primary.Attributes["listener_id"])

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if d.VServerId == vserverSet.VServerId && d.ULBId == lbSet.ULBId {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("lb ssl attachment not found"))
		})
	}
}

func testAccCheckLBSSLAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb_ssl_attachment" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeLBSSLAttachmentById(rs.Primary.Attributes["ssl_id"],
			rs.Primary.Attributes["load_balancer_id"],
			rs.Primary.Attributes["listener_id"])

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VServerId == rs.Primary.Attributes["listener_id"] || d.ULBId == rs.Primary.Attributes["load_balancer_id"] {
			return fmt.Errorf("lb ssl attachment still exists")
		}
	}

	return nil
}

const testAccLBSSLAttachmentConfig = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-ssl-attachment"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	name             = "tf-acc-lb-ssl-attachment"
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "https"
}

resource "ucloud_lb_ssl" "foo" {
	name 		= "tf-acc-lb-ssl-attachment"
	private_key = "${file("test-fixtures/private.key")}"
	user_cert 	= "${file("test-fixtures/user.crt")}"
	ca_cert 	= "${file("test-fixtures/ca.crt")}"
}

resource "ucloud_lb_ssl_attachment" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	ssl_id      	 = "${ucloud_lb_ssl.foo.id}"
}
`
