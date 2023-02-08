provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}

provider "aws" {
  region = "eu-west-1"

  default_tags {
    tags = data.roo_tags.defaults.aws_tags
  }

  assume_role {
    role_arn     = data.roo_aws_account.current.terraform_deploy_role_arn
    session_name = "geopoiesis"
  }
}

data roo_market_entries "all" {}
data "roo_aws_account" "current" {}
data "roo_tags" "defaults" {}

provider "roo" {
  default_ownership_group = "the-go-team"
  default_env_name        = var.env_name
  default_shard_name      = var.shard
}
