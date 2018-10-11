# Specify the provider and access details
provider "ucloud" {
     region = "${var.region}"
}

resource "ucloud_vpc" "foo" {
    name        = "tf-example-vpc-01"
    tag         = "tf-example"
    cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_vpc" "bar" {
    name        = "tf-example-vpc-02"
    tag         = "tf-example"
    cidr_blocks = ["10.10.0.0/16"]
}

resource "ucloud_vpc_peering_connection" "connection" {
    vpc_id      = "${ucloud_vpc.foo.id}"
    peer_vpc_id = "${ucloud_vpc.bar.id}"
}
