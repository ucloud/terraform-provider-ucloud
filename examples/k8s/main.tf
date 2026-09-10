# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-acc-uk8s-cluster2"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
  name       = "tf-acc-uk8s-cluster2"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id       = ucloud_vpc.foo.id
  subnet_id    = ucloud_subnet.foo.id
  name         = "tfuk8so2i0909a2"
  service_cidr = "172.16.0.0/16"
  cni_mode     = "VPC"
  password     = var.password
  charge_type  = "dynamic"
  k8s_version  = "1.32.8"
  enable_external_api_server = true
  kube_proxy {
    mode = "iptables"
  }

  master {
    availability_zones = [
      var.zone,
      var.zone,
      var.zone,
    ]
    machine_type = "O"
    cpu          = 2
    memory       = 4096
    uhost_family = "o2i"
    image_id     = "uimage-1rgr0eytog7f"

    boot_disk_type = "cloud_rssd"
    boot_disk_size = 40
    data_disk_type = "cloud_rssd"
    data_disk_size = 20
  }

  # Initial Worker created with the cluster, in addition to the nodes below.
  # Changing this block replaces the entire cluster.
  worker {
    availability_zone = var.zone
    machine_type      = "O"
    cpu               = 2
    memory            = 4096
    count             = 1
    uhost_family      = "o2i"
    min_cpu_platform  = "Intel/EmeraldRapids"
    image_id          = "uimage-1rgr0eytog7f"

    boot_disk_type = "cloud_rssd"
    boot_disk_size = 40
    data_disk_type = "cloud_rssd"
    data_disk_size = 20
    net_capability = "Ultra"
    max_pods       = 110

    labels = {
      role = "bootstrap"
    }
  }
}

resource "ucloud_uk8s_node_group" "workers" {
  cluster_id        = ucloud_uk8s_cluster.foo.id
  subnet_id         = ucloud_subnet.foo.id
  name              = "workers"
  instance_type     = "o-basic-2"
  uhost_family      = "o2i"
  charge_type       = "dynamic"
  image_id          = "uimage-1rgr0eytog7f"
  availability_zone = var.zone
  boot_disk_type    = "cloud_rssd"
  data_disk_type    = "cloud_rssd"
  data_disk_size    = 100

  labels = {
    environment = "test"
    role        = "worker1"
  }

  taint {
    key    = "dedicated"
    value  = "test"
    effect = "NoSchedule"
  }
}

resource "ucloud_uk8s_node" "foo" {
  cluster_id        = ucloud_uk8s_cluster.foo.id
  node_group_id     = ucloud_uk8s_node_group.workers.id
  subnet_id         = ucloud_subnet.foo.id
  password          = var.password
  image_id          = "uimage-1rgr0eytog7f"
  instance_type     = "o-basic-2"
  charge_type       = "dynamic"
  availability_zone = var.zone
  boot_disk_type    = "cloud_rssd"
  count             = 2
}
