provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}

provider "aws" {
  region = var.region

  default_tags {
    tags = data.roo_tags.defaults.aws_tags
  }

  assume_role {
    role_arn     = data.roo_aws_account.current.terraform_deploy_role_arn
    session_name = "geopoiesis"
  }
}

provider "circleci" {}

provider "hopper" {}

data "roo_aws_account" "current" {}

data "roo_tags" "defaults" {}

provider "roo" {
  default_ownership_group = "prodeng"
  default_env_name        = var.env_name
  default_shard_name      = var.shard
}