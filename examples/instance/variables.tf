variable "region" {
  description = "The region to create resources in"
  default     = "cn-sh2"
}

variable "instance_password" {
  default = "wA123456"
}

variable "count" {
  default = "1"
}

variable "count_format" {
  default = "%02d"
}
