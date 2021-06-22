# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
	name        = "tf-acc-uk8s-cluster"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-uk8s-cluster"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
	vpc_id	 	 = ucloud_vpc.foo.id
	subnet_id	 = ucloud_subnet.foo.id
	name  	 	 = "tf-acc-uk8s-cluster-basic"
	service_cidr = "172.16.0.0/16"
	password     = var.password
	charge_type  = "dynamic"

	kube_proxy {
		mode = "ipvs"
	}

	master {
	  availability_zones = [
		data.ucloud_zones.default.zones[0].id,
		data.ucloud_zones.default.zones[0].id,
		data.ucloud_zones.default.zones[0].id,
      ]
	  instance_type = "n-basic-2"
  	}
}

resource "ucloud_uk8s_node" "foo" {
	cluster_id    = ucloud_uk8s_cluster.foo.id
	subnet_id	  = ucloud_subnet.foo.id
	password      = var.password
	instance_type = "n-basic-2"
	charge_type   = "dynamic"
	availability_zone = data.ucloud_zones.default.zones[0].id

	count = 2
}
