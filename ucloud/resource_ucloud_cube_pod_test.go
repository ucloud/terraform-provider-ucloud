package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"testing"
)

func TestAccUCloudCubePod_basic(t *testing.T) {
	var podExtendInfo cubePodExtendInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_cube_pod.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckCubePodDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccCubePodConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckCubePodExists("ucloud_cube_pod.foo", &podExtendInfo),
					testAccCheckCubePodAttributes(&podExtendInfo),
					resource.TestCheckResourceAttr("ucloud_cube_pod.foo", "name", "tf-acc-cube-pod-basic"),
				),
			},
			{
				Config: testAccCubePodConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckCubePodExists("ucloud_cube_pod.foo", &podExtendInfo),
					testAccCheckCubePodAttributes(&podExtendInfo),
					resource.TestCheckResourceAttr("ucloud_cube_pod.foo", "name", "tf-acc-cube-pod-update"),
				),
			},
		},
	})
}

func testAccCheckCubePodExists(n string, cubePodSet *cubePodExtendInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("cube pod id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeCubePodById(rs.Primary.ID)

		log.Printf("[INFO] cube pod id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*cubePodSet = *ptr
		return nil
	}
}

func testAccCheckCubePodAttributes(cubePodSet *cubePodExtendInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cubePodSet.CubeExtendInfo.CubeId == "" {
			return fmt.Errorf("cube pod id is empty")
		}
		return nil
	}
}

func testAccCheckCubePodDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_cube_pod" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeCubePodById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.CubeExtendInfo.CubeId != "" {
			return fmt.Errorf("cube pod still exist")
		}
	}

	return nil
}

const testAccCubePodConfig = `
data "ucloud_zones" "default" {
}
data "ucloud_vpcs" "default" {
	name_regex = "DefaultVPC"
}
data "ucloud_subnets" "default" {
	vpc_id = "${data.ucloud_vpcs.default.vpcs.0.id}"
}

resource "ucloud_cube_pod" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name  	 	  = "tf-acc-cube-pod-basic"
	tag           = "tf-acc"
	vpc_id        = "${data.ucloud_vpcs.default.vpcs.0.id}"
	subnet_id     = "${data.ucloud_subnets.default.subnets.0.id}"
	pod           = "${file("test-fixtures/cube_pod.yml")}"
}
`

const testAccCubePodConfigUpdate = `
data "ucloud_zones" "default" {
}

data "ucloud_vpcs" "default" {
	name_regex = "DefaultVPC"
}
data "ucloud_subnets" "default" {
	vpc_id = "${data.ucloud_vpcs.default.vpcs.0.id}"
}

resource "ucloud_cube_pod" "foo" {
  	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name  	 	  = "tf-acc-cube-pod-update"
	tag           = "tf-acc"
	vpc_id        = "${data.ucloud_vpcs.default.vpcs.0.id}"
	subnet_id     = "${data.ucloud_subnets.default.subnets.0.id}"
	pod           = "${file("test-fixtures/cube_pod.yml")}"
}
`
