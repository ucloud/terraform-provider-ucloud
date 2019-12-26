package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLBRule_basic(t *testing.T) {
	var lbSet ulb.ULBSet
	var vserverSet ulb.ULBVServerSet
	var instance uhost.UHostInstanceSet
	var backendSet ulb.ULBBackendSet
	var policySet ulb.ULBPolicySet
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBRuleDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLBRuleConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckLBAttachmentExists("ucloud_lb_attachment.foo", &lbSet, &vserverSet, &backendSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckLBRuleExists("ucloud_lb_rule.foo", &lbSet, &vserverSet, &backendSet, &policySet),
					testAccCheckLBRuleAttributes(&policySet),
					resource.TestCheckResourceAttr("ucloud_lb_rule.foo", "domain", "www.ucloud.cn"),
				),
			},

			{
				Config: testAccLBRuleConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckLBAttachmentExists("ucloud_lb_attachment.foo", &lbSet, &vserverSet, &backendSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckLBRuleExists("ucloud_lb_rule.foo", &lbSet, &vserverSet, &backendSet, &policySet),
					testAccCheckLBRuleAttributes(&policySet),
					resource.TestCheckResourceAttr("ucloud_lb_rule.foo", "path", "/foo"),
				),
			},
		},
	})
}

func testAccCheckLBRuleExists(n string, lbSet *ulb.ULBSet, vserverSet *ulb.ULBVServerSet, backendSet *ulb.ULBBackendSet, policySet *ulb.ULBPolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("LBRule id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describePolicyById(lbSet.ULBId, vserverSet.VServerId, rs.Primary.ID)

		log.Printf("[INFO] LBRule id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*policySet = *ptr
		return nil
	}
}

func testAccCheckLBRuleAttributes(policySet *ulb.ULBPolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.PolicyId == "" {
			return fmt.Errorf("LBRule id is empty")
		}
		return nil
	}
}

func testAccCheckLBRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb_rule" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describePolicyById(
			rs.Primary.Attributes["load_balancer_id"],
			rs.Primary.Attributes["listener_id"],
			rs.Primary.ID,
		)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.PolicyId != "" {
			return fmt.Errorf("LBRule still exist")
		}
	}

	return nil
}

const testAccLBRuleConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_lb" "foo" {
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "http"
}

resource "ucloud_instance" "foo"{
	name              = "tf-acc-lb"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
}

resource "ucloud_lb_attachment" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_id      = "${ucloud_instance.foo.id}"
	port             = 80
}

resource "ucloud_lb_rule" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	backend_ids      = ["${ucloud_lb_attachment.foo.id}"]
	domain           = "www.ucloud.cn"
}
`
const testAccLBRuleConfigTwo = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-rule"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "http"
}

resource "ucloud_instance" "foo"{
	name              = "tf-acc-lb-rule"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
}

resource "ucloud_lb_attachment" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_id      = "${ucloud_instance.foo.id}"
	port             = 80
}

resource "ucloud_lb_rule" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	backend_ids      = ["${ucloud_lb_attachment.foo.id}"]
	path             = "/foo"
}
`
