package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudLBAttachmentsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBAttachmentsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_attachments.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_attachments.foo", "lb_attachments.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_lb_attachments.foo", "lb_attachments.0.port", "80"),
				),
			},
		},
	})
}

const testAccDataLBAttachmentsConfig = `

variable "name" {
	default = "tf-acc-lb-attachments-dataSource-basic"
}

variable "count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_lb" "foo" {
	name = "${var.name}"
	tag  = "tf-acc"
}

resource "ucloud_lb_listener" "foo" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	protocol         = "http"
}

resource "ucloud_instance" "foo"{
	name              = "${var.name}-${format(var.count_format, count.index+1)}"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
	count 			  = "${var.count}"
}

resource "ucloud_lb_attachment" "foo" {
	count 			 = "${var.count}"
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_id      = "${element(ucloud_instance.foo.*.id, count.index)}"
	port             = 80
}

data "ucloud_lb_attachments" "foo" {
	ids 			 = ["${ucloud_lb_attachment.foo.*.id}"]
	listener_id      = "${ucloud_lb_listener.foo.id}"
	load_balancer_id = "${ucloud_lb.foo.id}"
}
`
