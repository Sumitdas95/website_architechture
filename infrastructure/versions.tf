terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.32.0"
    }
    circleci = {
      source = "terraform-registry.deliveroo.net/deliveroo/circleci"
      version = "~> 1.0"
    }
    datadog = {
      source = "DataDog/datadog"
    }
    hopper = {
      source  = "terraform-registry.deliveroo.net/deliveroo/hopper"
      version = "~> 1.0"
    }
    sentry = {
      source  = "terraform-registry.deliveroo.net/deliveroo/sentry"
      version = "~> 1.0"
    }
    roo = {
      source  = "terraform-registry.deliveroo.net/deliveroo/roo"
      version = "~> 1.0"
    }
  }
  required_version = "~> 1.1.0"
}
