package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudLBListenersDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBListenersConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_listeners.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_listeners.foo", "lb_listeners.#", "2"),
				),
			},
		},
	})
}

const testAccDataLBListenersConfig = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-listener"
	tag  = "tf-acc"
}

variable "count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

resource "ucloud_lb_listener" "foo" {
	count = "${var.count}"
	load_balancer_id  = "${ucloud_lb.foo.id}"
	protocol          = "https"
	port			  = 80+"${count.index+1}"
	method            = "source"
	name              = "tf-acc-lb-listeners-${format(var.count_format, count.index+1)}"
	path              = "/foo"
	idle_timeout      = 80
	persistence_type  = "server_insert"
	health_check_type = "path"
}

data "ucloud_lb_listeners" "foo" {
	ids = ["${ucloud_lb_listener.foo.*.id}"]
	load_balancer_id = "${ucloud_lb.foo.id}"
}
`
