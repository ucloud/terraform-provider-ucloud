package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLBListener_basic(t *testing.T) {
	var lbSet ulb.ULBSet
	var vserverSet ulb.ULBVServerSet
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb_listener.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBListenerDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckLBListenerAttributes(&vserverSet),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "protocol", "https"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "method", "source"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "name", "tf-acc-lb-listener"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "idle_timeout", "80"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "persistence_type", "server_insert"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "health_check_type", "path"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "path", "/foo"),
				),
			},

			{
				Config: testAccLBListenerConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBListenerExists("ucloud_lb_listener.foo", &lbSet, &vserverSet),
					testAccCheckLBListenerAttributes(&vserverSet),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "protocol", "https"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "method", "roundrobin"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "name", "tf-acc-lb-listener-two"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "idle_timeout", "100"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "persistence_type", "none"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "health_check_type", "port"),
					resource.TestCheckResourceAttr("ucloud_lb_listener.foo", "domain", "www.ucloud.cn"),
				),
			},
		},
	})
}

func testAccCheckLBListenerExists(n string, lbSet *ulb.ULBSet, vserverSet *ulb.ULBVServerSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("LBListener id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVServerById(lbSet.ULBId, rs.Primary.ID)

		log.Printf("[INFO] LBListener id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*vserverSet = *ptr
		return nil
	}
}

func testAccCheckLBListenerAttributes(vserverSet *ulb.ULBVServerSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if vserverSet.VServerId == "" {
			return fmt.Errorf("LBListener id is empty")
		}
		return nil
	}
}

func testAccCheckLBListenerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb_listener" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVServerById(rs.Primary.Attributes["load_balancer_id"], rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VServerId != "" {
			return fmt.Errorf("LBListener still exist")
		}
	}

	return nil
}

const testAccLBListenerConfig = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-listener"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id  = "${ucloud_lb.foo.id}"
	protocol          = "https"
	method            = "source"
	name              = "tf-acc-lb-listener"
	path              = "/foo"
	idle_timeout      = 80
	persistence_type  = "server_insert"
	health_check_type = "path"
}
`

const testAccLBListenerConfigTwo = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-listener"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id  = "${ucloud_lb.foo.id}"
	protocol          = "https"
	method            = "roundrobin"
	name              = "tf-acc-lb-listener-two"
	idle_timeout      = 100
	persistence_type  = "none"
	health_check_type = "port"
	domain            = "www.ucloud.cn"
}
`
