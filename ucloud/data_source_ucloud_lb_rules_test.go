package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudLBRulesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBRulesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_rules.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_rules.foo", "lb_rules.#", "2"),
				),
			},
		},
	})
}

const testAccDataLBRulesConfig = `
variable "instance_count" {
	default = 2
}

variable "name" {
	default = "tf-acc-lb-rules-dataSource-basic"
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
	count 			  = "${var.instance_count}"
}

resource "ucloud_lb_attachment" "foo" {
	count 			 = "${var.instance_count}"
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	resource_id      = ucloud_instance.foo[count.index].id
	port             = 80
}

resource "ucloud_lb_rule" "test_domain" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	backend_ids      = ucloud_lb_attachment.foo.*.id
	domain           = "www.ucloud.cn"
}

resource "ucloud_lb_rule" "test_path" {
	load_balancer_id = "${ucloud_lb.foo.id}"
	listener_id      = "${ucloud_lb_listener.foo.id}"
	backend_ids      = ucloud_lb_attachment.foo.*.id
	path             = "/foo"
}


data "ucloud_lb_rules" "foo" {
	ids 			 = [ucloud_lb_rule.test_domain.id, ucloud_lb_rule.test_path.id]
	listener_id      = "${ucloud_lb_listener.foo.id}"
	load_balancer_id = "${ucloud_lb.foo.id}"
}
`
