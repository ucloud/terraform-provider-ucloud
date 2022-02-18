variable "region" {
  default = "cn-bj2"
}

variable "zone" {
  default = "cn-bj2-05"
}

variable "redis_password" {
  default = "2018_UClou"
}

variable "charge_type" {
  description = "charge type"
  type = string
  default = "month"
}
