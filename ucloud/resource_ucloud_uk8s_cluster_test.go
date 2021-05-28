package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"log"
	"testing"
)

func TestAccUCloudUK8sCluster_basic(t *testing.T) {
	var uk8sClusterSet uk8s.ClusterSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_uk8s_cluster.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUK8sClusterDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUK8sClusterConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8sClusterExists("ucloud_uk8s_cluster.foo", &uk8sClusterSet),
					testAccCheckUK8sClusterAttributes(&uk8sClusterSet),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic"),
				),
			},
			{
				Config: testAccUK8sClusterConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8sClusterExists("ucloud_uk8s_cluster.foo", &uk8sClusterSet),
					testAccCheckUK8sClusterAttributes(&uk8sClusterSet),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic-update"),
				),
			},
		},
	})
}

func testAccCheckUK8sClusterExists(n string, uk8sClusterSet *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUK8sClusterById(rs.Primary.ID)

		log.Printf("[INFO] uk8s cluster id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*uk8sClusterSet = *ptr
		return nil
	}
}

func testAccCheckUK8sClusterAttributes(uk8sClusterSet *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if uk8sClusterSet.ClusterId == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}
		return nil
	}
}

func testAccCheckUK8sClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_uk8s_cluster" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUK8sClusterById(rs.Primary.ID)

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

const testAccUK8sClusterConfig = `
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
	master_instance_type = "n-basic-2"
   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}

	nodes {
	  instance_type = "n-basic-2"
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
}
`

const testAccUK8sClusterConfigUpdate = `
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
	master_instance_type = "n-basic-2"

   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
   	master {
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}

	nodes {
	  instance_type = "n-basic-2"
	  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  	}
}
`
