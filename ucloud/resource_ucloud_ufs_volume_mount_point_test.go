package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"log"
	"testing"
)

func TestAccUCloudUFSVolumeMountPoint_basic(t *testing.T) {
	var mountPointSet ufs.MountPointInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_ufs_volume_mount_point.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUFSVolumeMountPointDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUFSVolumeMountPointConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUFSVolumeMountPointExists("ucloud_ufs_volume_mount_point.foo", &mountPointSet),
					resource.TestCheckResourceAttr("ucloud_ufs_volume_mount_point.foo", "name", "tf-acc-ufs-mount-point-basic"),
				),
			},
		},
	})
}

func testAccCheckUFSVolumeMountPointExists(n string, ufsSet *ufs.MountPointInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ufs volume mount point id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		volumeId := rs.Primary.Attributes["volume_id"]
		vpcId := rs.Primary.Attributes["vpc_id"]
		subnetId := rs.Primary.Attributes["subnet_id"]

		ptr, err := client.describeUFSVolumeMountPointById(volumeId, vpcId, subnetId)

		log.Printf("[INFO] ufs volume mount point id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*ufsSet = *ptr
		return nil
	}
}

func testAccCheckUFSVolumeMountPointDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_ufs_volume_mount_point" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		volumeId := rs.Primary.Attributes["volume_id"]
		vpcId := rs.Primary.Attributes["vpc_id"]
		subnetId := rs.Primary.Attributes["subnet_id"]
		d, err := client.describeUFSVolumeMountPointById(volumeId, vpcId, subnetId)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}
		if d != nil {
			return fmt.Errorf("ufs volume mount point still exist")
		}
	}

	return nil
}

const testAccUFSVolumeMountPointConfig = `
resource "ucloud_vpc" "default" {
  name        = "tf-acc-mount-point-size"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.128.0/17"]
}

resource "ucloud_subnet" "default" {
  name       = "tf-acc-ufs-mount-point-size"
  tag        = "tf-acc"
  cidr_block = "192.168.128.0/17"
  vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_ufs_volume" "default" {
	name  	 	  = "tf-acc-ufs-mount-point-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}

resource "ucloud_ufs_volume_mount_point" "foo" {
	name  	 	  = "tf-acc-ufs-mount-point-basic"
	volume_id 	  = "${ucloud_ufs_volume.default.id}"
  	vpc_id        = "${ucloud_vpc.default.id}"
  	subnet_id     = "${ucloud_subnet.default.id}"
}
`
