package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudLabelAttachment_basic(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_label_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLabelDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLabelAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_label.foo", "key", "tf-acc-label-key"),
					resource.TestCheckResourceAttr("ucloud_label.foo", "value", "tf-acc-label-value"),
					resource.TestCheckResourceAttr("ucloud_label_attachment.foo", "key", "tf-acc-label-key"),
					resource.TestCheckResourceAttr("ucloud_label_attachment.foo", "value", "tf-acc-label-value"),
				),
			},
		},
	})
}

const testAccLabelAttachmentConfig = `
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
	name  	 	= "tf-acc-vip-basic-update"
	remark 		= "test-update"
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

`
