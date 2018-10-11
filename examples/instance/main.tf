# Specify the provider and access details
provider "ucloud" {
    region = "${var.region}"
}

# Query availability zone
data "ucloud_zones" "default" {
}

# Query image
data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

# Create security group
resource "ucloud_security_group" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # HTTP access from LAN
    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }

    # HTTPS access from LAN
    rules {
        port_range = "443"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

# Create vpc
resource "ucloud_vpc" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # vpc network
    cidr_blocks = ["192.168.0.0/16"]
}

# Create subnet
resource "ucloud_subnet" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"
 
    # subnet's network must be contained by vpc network
    # and a subnet must have least 8 ip addresses in it (netmask < 30).
    cidr_block = "192.168.1.0/24"
    vpc_id     = "${ucloud_vpc.default.id}"
}

# Create a web server
resource "ucloud_instance" "web" {
    name              = "tf-example-instance-${format(var.count_format, count.index+1)}"
    tag               = "tf-example"
    availability_zone = "${data.ucloud_zones.default.zones.0.id}"
    image_id          = "${data.ucloud_images.default.images.0.id}"
    instance_type     = "n-standard-1"

    # use cloud disk as data disk
    data_disk_size     = 50
    data_disk_category = "Disk"
    root_password      = "${var.instance_password}"

    # we will put all the instances into same vpc and subnet,
    # so they can communicate with each other.
    vpc_id    = "${ucloud_vpc.default.id}"
    subnet_id = "${ucloud_subnet.default.id}"

    # this ecurity group to allow HTTP and HTTPS access
    security_group = "${ucloud_security_group.default.id}"

    count = "${var.count}"
}
