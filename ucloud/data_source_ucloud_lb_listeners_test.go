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
					resource.TestCheckResourceAttr("data.ucloud_lb_listeners.foo", "lb_listeners.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_lb_listeners.foo", "lb_listeners.0.port", "80"),
				),
			},
		},
	})
}

func TestAccUCloudLBListenersDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBListenersConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_listeners.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_listeners.foo", "lb_listeners.#", "2"),
				),
			},
		},
	})
}

const testAccDataLBListenersConfig = `
variable "name" {
	default = "tf-acc-lb-listeners-dataSource-basic"
}

resource "ucloud_lb" "foo" {
	name = "${var.name}"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id  = "${ucloud_lb.foo.id}"
	protocol          = "https"
	port			  = "80"
	method            = "source"
	name              = "${var.name}"
	path              = "/foo"
	idle_timeout      = 80
	persistence_type  = "server_insert"
	health_check_type = "path"
}

data "ucloud_lb_listeners" "foo" {	
	load_balancer_id = "${ucloud_lb.foo.id}"
	name_regex  = "${ucloud_lb_listener.foo.name}"
}
`

const testAccDataLBListenersConfigIds = `

variable "name" {
	default = "tf-acc-lb-listeners-dataSource-ids"
}

resource "ucloud_lb" "foo" {
	name = "${var.name}"
	tag  = "tf-acc"
}

variable "count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

variable "port_range" {
	default = {
		"0" = 80
		"1" = 88
	}
}

resource "ucloud_lb_listener" "foo" {
	count 			  = "${var.count}"
	load_balancer_id  = "${ucloud_lb.foo.id}"
	protocol          = "https"
	port			  = "${var.port_range[count.index]}"
	method            = "source"
	name              = "${var.name}-${format(var.count_format, count.index+1)}"
	path              = "/foo"
	idle_timeout      = 80
	persistence_type  = "server_insert"
	health_check_type = "path"
}

data "ucloud_lb_listeners" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	ids				 = ["${ucloud_lb_listener.foo.*.id}"]
}
`
