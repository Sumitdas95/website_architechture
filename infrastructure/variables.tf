variable "datadog_api_key" {}
variable "datadog_app_key" {}

variable "shard" {}
variable "env_name" {}

variable "region" {
  default = "eu-west-1"
}

variable "terraform_repo" {
  default = "test-sonarqube"
}

locals {
  TIER          = "3"
  TEAM_NAME     = "the-go-team"
  SLACK_CHANNEL = "https://deliveroo.slack.com/archives/C8NSBPPS4"
  ROTA          = ""
  OWNER_EMAIL   = "consumer-eng@deliveroo.co.uk"

  REPO_NAME     = "test-sonarqube"
  PLAYBOOK_LINK = "https://github.com/deliveroo/test-sonarqube/blob/main/PLAYBOOK.md"

  HOPPER_APP_NAME        = "test-sonarqube"
  HOPPER_APP_DESCRIPTION = "A sample service for sharding in Go!"
  HOPPER_PATH            = ".hopper/config.yml"

  DATADOG_ALERT_CHANNELS = {
    production = "@slack-test-sonarqube-alerts"
    staging    = "@slack-test-sonarqube-staging"
  }
}

# Supportive locals for creating monitors and dashboards within Datadog. Useful
# for specifying options, defaults and possible filter values.
locals {
  var_env     = "env"
  var_cluster = "clustername"
  var_shard   = "shard"

  production = "production"
  staging    = "staging"

  cluster_production = "production"
  cluster_staging    = "staging"

  shards_available = [
    "global",
    "apc-01",
    "eur-01",
    "fra-01",
    "gbr-01",
    "mle-01"
  ]
}