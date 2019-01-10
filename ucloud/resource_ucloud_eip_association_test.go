package ucloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudEIPAssociation_basic(t *testing.T) {
	var eip unet.UnetEIPSet
	var instance uhost.UHostInstanceSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_eip_association.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPAssociationDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckEIPAssociationExists("ucloud_eip_association.foo", &eip, &instance),
					resource.TestCheckResourceAttr("ucloud_eip_association.foo", "resource_type", "instance"),
				),
			},
		},
	})
}

func testAccCheckEIPAssociationExists(n string, eip *unet.UnetEIPSet, instance *uhost.UHostInstanceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("eip association id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)

		eipId := rs.Primary.Attributes["eip_id"]
		resourceId := rs.Primary.Attributes["resource_id"]

		return resource.Retry(3*time.Minute, func() *resource.RetryError {
			d, err := client.describeEIPResourceById(eipId, resourceId)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if d.ResourceId == instance.UHostId {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("eip association not found"))
		})
	}
}

func testAccCheckEIPAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_eip_association" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeEIPById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.Resource.ResourceId == rs.Primary.Attributes["resource_id"] {
			return fmt.Errorf("eip associatoin still exists")
		}
	}

	return nil
}

const testAccEIPAssociationConfig = `
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        =  "base"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-eip-association-eip"
	tag           = "tf-acc"
	internet_type = "bgp"
	bandwidth     = 1
	duration      = 1
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-eip-association-intance"
	tag               = "tf-acc"
	instance_type     = "n-highcpu-1"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	root_password     = "wA123456"
	duration          = 1
	charge_type       = "month"
}

resource "ucloud_eip_association" "foo" {
	eip_id        = "${ucloud_eip.foo.id}"
	resource_id   = "${ucloud_instance.foo.id}"
}
`
