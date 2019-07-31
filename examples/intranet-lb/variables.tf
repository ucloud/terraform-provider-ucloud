variable "region" {
  description = "The region to create resources in"
  default     = "cn-bj2"
}

variable "instance_password" {
  default = "wA123456"
}

variable "instance_count" {
  default = 2
}

variable "count_format" {
  default = "%02d"
}

