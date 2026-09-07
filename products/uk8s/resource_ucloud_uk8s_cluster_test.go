package uk8s_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
)

func TestAccUCloudUK8SCluster_basic(t *testing.T) {
	var cluster uk8s.ClusterSet

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
					testAccCheckUK8SClusterExists("ucloud_uk8s_cluster.foo", &cluster),
					testAccCheckUK8SClusterAttributes(&cluster),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic"),
				),
			},
			{
				Config: testAccUK8SClusterConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8SClusterExists("ucloud_uk8s_cluster.foo", &cluster),
					testAccCheckUK8SClusterAttributes(&cluster),
					resource.TestCheckResourceAttr("ucloud_uk8s_cluster.foo", "name", "tf-acc-uk8s-cluster-basic-update"),
				),
			},
		},
	})
}

func testAccCheckUK8SClusterExists(name string, target *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}

		client, err := testAccUK8SClient()
		if err != nil {
			return err
		}
		cluster, found, err := describeAccUK8SClusterByID(client, item.Primary.ID)
		log.Printf("[INFO] uk8s cluster id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("uk8s cluster %q is not found", item.Primary.ID)
		}
		*target = *cluster
		return nil
	}
}

func testAccCheckUK8SClusterAttributes(cluster *uk8s.ClusterSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if cluster.ClusterId == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}
		return nil
	}
}

func testAccCheckUK8SClusterDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_uk8s_cluster" {
			continue
		}

		client, err := testAccUK8SClient()
		if err != nil {
			return err
		}
		cluster, found, err := describeAccUK8SClusterByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if cluster.ClusterId != "" {
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
    vpc_id       = "${ucloud_vpc.foo.id}"
    subnet_id    = "${ucloud_subnet.foo.id}"
    name         = "tf-acc-uk8s-cluster-basic"
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
    vpc_id       = "${ucloud_vpc.foo.id}"
    subnet_id    = "${ucloud_subnet.foo.id}"
    name         = "tf-acc-uk8s-cluster-basic-update"
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
