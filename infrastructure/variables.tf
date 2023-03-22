variable "shard" {}

variable "env_name" {
  description = "Name of the hosting environment"
  type        = string
  default     = "sandbox"
}

variable "datadog_api_key" {
  description = "DataDog api key"
  type        = string
}

variable "datadog_app_key" {
  description = "DataDog app key"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "eu-west-1"
}