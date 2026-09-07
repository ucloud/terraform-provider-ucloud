package ufs_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"log"
	"testing"
)

func TestAccUCloudUFSVolume_basic(t *testing.T) {
	var ufsSet ufs.UFSVolumeInfo2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_ufs_volume.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUFSVolumeDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUFSVolumeConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUFSVolumeExists("ucloud_ufs_volume.foo", &ufsSet),
					testAccCheckUFSVolumeAttributes(&ufsSet),
					resource.TestCheckResourceAttr("ucloud_ufs_volume.foo", "name", "tf-acc-ufs-basic"),
					resource.TestCheckResourceAttr("ucloud_ufs_volume.foo", "size", "500"),
				),
			},
			{
				Config: testAccUFSVolumeConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUFSVolumeExists("ucloud_ufs_volume.foo", &ufsSet),
					testAccCheckUFSVolumeAttributes(&ufsSet),
					resource.TestCheckResourceAttr("ucloud_ufs_volume.foo", "name", "tf-acc-ufs-basic"),
					resource.TestCheckResourceAttr("ucloud_ufs_volume.foo", "size", "600"),
				),
			},
		},
	})
}

func testAccCheckUFSVolumeExists(n string, ufsSet *ufs.UFSVolumeInfo2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ufs volume id is empty")
		}

		client, err := testAccUFSClient()
		if err != nil {
			return err
		}
		ptr, found, err := describeAccUFSVolumeByID(client, rs.Primary.ID)

		log.Printf("[INFO] ufs volume id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("ufs volume %q is not found", rs.Primary.ID)
		}

		*ufsSet = *ptr
		return nil
	}
}

func testAccCheckUFSVolumeAttributes(ufsSet *ufs.UFSVolumeInfo2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ufsSet.VolumeId == "" {
			return fmt.Errorf("ufs volume id is empty")
		}
		return nil
	}
}

func testAccCheckUFSVolumeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_ufs_volume" {
			continue
		}

		client, err := testAccUFSClient()
		if err != nil {
			return err
		}
		d, found, err := describeAccUFSVolumeByID(client, rs.Primary.ID)

		if err != nil {
			return err
		}
		if !found {
			continue
		}

		if d.VolumeId != "" {
			return fmt.Errorf("ufs volume still exist")
		}
	}

	return nil
}

const testAccUFSVolumeConfig = `
resource "ucloud_ufs_volume" "foo" {
	name  	 	  = "tf-acc-ufs-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}
`

const testAccUFSVolumeConfigUpdate = `
resource "ucloud_ufs_volume" "foo" {
	name  	 	  = "tf-acc-ufs-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 600
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}
`
