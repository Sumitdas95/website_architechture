terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    circleci = {
      source = "terraform-registry.deliveroo.net/deliveroo/circleci"
    }
    datadog = {
      source = "DataDog/datadog"
    }
    hopper = {
      source  = "terraform-registry.deliveroo.net/deliveroo/hopper"
      version = "~> 1.0"
    }
    random = {
      source = "hashicorp/random"
    }
    roo = {
      source  = "terraform-registry.deliveroo.net/deliveroo/roo"
      version = "~> 1.0"
    }
  }
}