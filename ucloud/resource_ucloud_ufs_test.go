package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"log"
	"testing"
)

func TestAccUCloudUFS_basic(t *testing.T) {
	var ufsSet ufs.UFSVolumeInfo2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_ufs.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUFSDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUFSConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUFSExists("ucloud_ufs.foo", &ufsSet),
					testAccCheckUFSAttributes(&ufsSet),
					resource.TestCheckResourceAttr("ucloud_ufs.foo", "name", "tf-acc-ufs-basic"),
					resource.TestCheckResourceAttr("ucloud_ufs.foo", "size", "500"),
				),
			},
			{
				Config: testAccUFSConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUFSExists("ucloud_ufs.foo", &ufsSet),
					testAccCheckUFSAttributes(&ufsSet),
					resource.TestCheckResourceAttr("ucloud_ufs.foo", "name", "tf-acc-ufs-basic"),
					resource.TestCheckResourceAttr("ucloud_ufs.foo", "size", "600"),
				),
			},
		},
	})
}

func testAccCheckUFSExists(n string, ufsSet *ufs.UFSVolumeInfo2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ufs id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUFSById(rs.Primary.ID)

		log.Printf("[INFO] ufs id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*ufsSet = *ptr
		return nil
	}
}

func testAccCheckUFSAttributes(ufsSet *ufs.UFSVolumeInfo2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ufsSet.VolumeId == "" {
			return fmt.Errorf("ufs id is empty")
		}
		return nil
	}
}

func testAccCheckUFSDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_ufs" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUFSById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VolumeId != "" {
			return fmt.Errorf("ufs still exist")
		}
	}

	return nil
}

const testAccUFSConfig = `
resource "ucloud_ufs" "foo" {
	name  	 	  = "tf-acc-ufs-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}
`

const testAccUFSConfigUpdate = `
resource "ucloud_ufs" "foo" {
	name  	 	  = "tf-acc-ufs-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 600
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}
`
