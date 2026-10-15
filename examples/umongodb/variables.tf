variable "region" {
  default = "cn-bj2"
}

variable "zone" {
  default = "cn-bj2-05"
}

variable "db_version" {
  default = "MongoDB 4.2"
}

variable "machine_type_id" {
  default = "o.mongo2m.medium"
}

variable "admin_password" {
  type = string
}
