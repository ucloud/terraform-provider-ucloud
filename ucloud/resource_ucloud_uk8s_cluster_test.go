package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"log"
	"testing"
)

func TestAccUCloudUK8SCluster_basic(t *testing.T) {
	var uk8sClusterSet uk8s.ClusterSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_uk8s_cluster.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUK8SClusterDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUK8SClusterConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8SClusterExists("ucloud_uk8s_cluster.foo", &uk8sClusterSet),
					testAccCheckUK8SClusterAttributes(&uk8sClusterSet),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic"),
				),
			},
			{
				Config: testAccUK8SClusterConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8SClusterExists("ucloud_uk8s_cluster.foo", &uk8sClusterSet),
					testAccCheckUK8SClusterAttributes(&uk8sClusterSet),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic-update"),
				),
			},
		},
	})
}

func testAccCheckUK8SClusterExists(n string, uk8sClusterSet *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUK8SClusterById(rs.Primary.ID)

		log.Printf("[INFO] uk8s cluster id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*uk8sClusterSet = *ptr
		return nil
	}
}

func testAccCheckUK8SClusterAttributes(uk8sClusterSet *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if uk8sClusterSet.ClusterId == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}
		return nil
	}
}

func testAccCheckUK8SClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_uk8s_cluster" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUK8SClusterById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.ClusterId != "" {
			return fmt.Errorf("uk8s cluster still exist")
		}
	}

	return nil
}

const testAccUK8SClusterConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-uk8s-cluster"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-uk8s-cluster"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
	vpc_id	 	 = "${ucloud_vpc.foo.id}"
	subnet_id	 = "${ucloud_subnet.foo.id}"
	name  	 	 = "tf-acc-uk8s-cluster-basic"
	service_cidr = "172.16.0.0/16"
	password     = "ucloud_2021"
	charge_type  = "dynamic"

	master {
	  availability_zones = [
		"${data.ucloud_zones.default.zones.0.id}",
		"${data.ucloud_zones.default.zones.0.id}",
		"${data.ucloud_zones.default.zones.0.id}",
      ]
	  instance_type = "n-basic-2"
  	}
}
`

const testAccUK8SClusterConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-uk8s-cluster"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-uk8s-cluster"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
	vpc_id	 	 = "${ucloud_vpc.foo.id}"
	subnet_id	 = "${ucloud_subnet.foo.id}"
	name  	 	 = "tf-acc-uk8s-cluster-basic-update"
	service_cidr = "172.16.0.0/16"
	password     = "ucloud_2021"
	charge_type  = "dynamic"

   	master {
	  availability_zones = [
		"${data.ucloud_zones.default.zones.0.id}",
		"${data.ucloud_zones.default.zones.0.id}",
		"${data.ucloud_zones.default.zones.0.id}",
      ]
	  instance_type = "n-basic-2"
  	}
}
`
