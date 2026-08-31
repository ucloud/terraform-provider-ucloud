package unet_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudSecurityGroup_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var securityGroup unet.FirewallDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_security_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecurityGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &securityGroup),
					testAccCheckSecurityGroupAttributes(&securityGroup),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", fmt.Sprintf("tf-acc-security-group-%d", rInt)),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.port_range", "80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.protocol", "tcp"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.cidr_block", "192.168.0.0/16"),
				),
			},

			{
				Config: testAccSecurityGroupConfigTwo(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &securityGroup),
					testAccCheckSecurityGroupAttributes(&securityGroup),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", fmt.Sprintf("tf-acc-security-group-%d-two", rInt)),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "tag", "Default"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.port_range", "20-80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.protocol", "tcp"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.cidr_block", "0.0.0.0/0"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupExists(name string, securityGroup *unet.FirewallDataSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("security group id is empty")
		}

		client, err := testAccUNetClient()
		if err != nil {
			return err
		}
		pointer, found, err := describeAccSecurityGroupByID(client, item.Primary.ID)
		log.Printf("[INFO] security group id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("security group %q is not found", item.Primary.ID)
		}
		*securityGroup = *pointer
		return nil
	}
}

func testAccCheckSecurityGroupAttributes(securityGroup *unet.FirewallDataSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if securityGroup.FWId == "" {
			return fmt.Errorf("security group id is empty")
		}
		return nil
	}
}

func testAccSecurityGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_security_group" "foo" {
	name = "tf-acc-security-group-%d"
	tag  = "tf-acc"
	rules {
		port_range = "80"
		protocol   = "tcp"
		cidr_block = "192.168.0.0/16"
		policy     = "accept"
		priority   = "high"
	}
}`, rInt)
}

func testAccSecurityGroupConfigTwo(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_security_group" "foo" {
	name = "tf-acc-security-group-%d-two"
	tag  = ""
	rules {
		port_range = "20-80"
		protocol   = "tcp"
		cidr_block = "0.0.0.0/0"
		policy     = "accept"
		priority   = "high"
	}
}`, rInt)
}
