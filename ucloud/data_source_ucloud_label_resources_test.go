package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudLabelResourcesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLabelResourcesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_labels.foo"),
					resource.TestCheckResourceAttr("data.ucloud_label_resources.foo", "resources.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_label_resources.foo", "resources.0.name", "tf-acc-vip-basic"),
					resource.TestCheckResourceAttr("data.ucloud_label_resources.foo", "resources.0.type", "vip"),
				),
			},
		},
	})
}

const testAccDataLabelResourcesConfig = `

resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vip"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vip"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}
resource "ucloud_vip" "foo" {
	vpc_id	 	= "${ucloud_vpc.foo.id}"
	subnet_id	= "${ucloud_subnet.foo.id}"
	name  	 	= "tf-acc-vip-basic"
	remark 		= "test"
	tag         = "tf-acc"
}
resource "ucloud_label" "foo" {
	key   = "tf-acc-label-key"
	value = "tf-acc-label-value"
}
resource "ucloud_label_attachment" "foo" {
	key   = "${ucloud_label.foo.key}"
	value = "${ucloud_label.foo.value}"
    resource = "${ucloud_vip.foo.id}"
}

data "ucloud_labels" "foo" {
	key_regex   = "^${ucloud_label_attachment.foo.key}$"
}

data "ucloud_label_resources" "foo" {
	key   = "${ucloud_label_attachment.foo.key}"
	value =  "${ucloud_label_attachment.foo.value}"
	resource_types = ["vip"]
	project_ids = ["${data.ucloud_labels.foo.labels.0.projects.0.id}"]
}
`
