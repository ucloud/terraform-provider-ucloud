package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLBAttachment_basic(t *testing.T) {
	var lbSet ulb.ULBSet
	var vserverSet ulb.ULBVServerSet
	var instance uhost.UHostInstanceSet
	var backendSet ulb.ULBBackendSet
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBAttachmentDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccLBAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckLBAttachmentExists("ucloud_lb_attachment.foo", &lbSet, &vserverSet, &backendSet),
					testAccCheckLBAttachmentAttributes(&backendSet),
					resource.TestCheckResourceAttr("ucloud_lb_attachment.foo", "port", "80"),
				),
			},

			resource.TestStep{
				Config: testAccLBAttachmentConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckLBAttachmentExists("ucloud_lb_attachment.foo", &lbSet, &vserverSet, &backendSet),
					testAccCheckLBAttachmentAttributes(&backendSet),
					resource.TestCheckResourceAttr("ucloud_lb_attachment.foo", "port", "1080"),
				),
			},
		},
	})
}

func testAccCheckLBAttachmentExists(n string, lbSet *ulb.ULBSet, vserverSet *ulb.ULBVServerSet, backendSet *ulb.ULBBackendSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("LBAttachment id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeBackendById(lbSet.ULBId, vserverSet.VServerId, rs.Primary.ID)

		log.Printf("[INFO] LBAttachment id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*backendSet = *ptr
		return nil
	}
}

func testAccCheckLBAttachmentAttributes(backendSet *ulb.ULBBackendSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if backendSet.BackendId == "" {
			return fmt.Errorf("LBAttachment id is empty")
		}
		return nil
	}
}

func testAccCheckLBAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb_attachment" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeBackendById(
			rs.Primary.Attributes["load_balancer_id"],
			rs.Primary.Attributes["listener_id"],
			rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.BackendId != "" {
			return fmt.Errorf("LBAttachment still exist")
		}
	}

	return nil
}

const testAccLBAttachmentConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-attachment"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	name             = "tf-acc-lb-attachment"
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "https"
}

resource "ucloud_instance" "foo"{
	name              = "tf-acc-lb-attachment"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
}

resource "ucloud_lb_attachment" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_type    = "instance"
	resource_id      = "${ucloud_instance.foo.id}"
	port             = 80
}
`
const testAccLBAttachmentConfigTwo = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-attachment"
	tag  = "tf-acc"
}


resource "ucloud_lb_listener" "foo" {
	name             = "tf-acc-lb-attachment"
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "https"
}

resource "ucloud_instance" "foo"{
	name              = "tf-acc-lb-attachment"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
}

resource "ucloud_lb_attachment" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_type    = "instance"
	resource_id      = "${ucloud_instance.foo.id}"
	port             = 1080
}
`
