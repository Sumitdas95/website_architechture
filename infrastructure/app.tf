module "test-sonarqube" {
  source  = "terraform-registry.deliveroo.net/deliveroo/roo_app_basic/aws"
  version = "~> 4.0"

  # Only deploy the sentry creation in the global shard
  count = var.shard == "global" ? 1 : 0

  app_name      = local.HOPPER_APP_NAME
  description   = local.HOPPER_APP_DESCRIPTION
  owner_email   = local.OWNER_EMAIL
  playbook_link = local.PLAYBOOK_LINK
  repo_name     = local.REPO_NAME
  rota_link     = local.ROTA
  slack_url     = local.SLACK_CHANNEL
  team_name     = local.TEAM_NAME
  tier          = local.TIER

  default_cooldown = 3

  sentry_enabled     = true
  uses_feature_flags = true

  datadog_alert_target = {
    production = local.DATADOG_ALERT_CHANNELS.production
    staging    = local.DATADOG_ALERT_CHANNELS.staging
  }

  yaml_path = local.HOPPER_PATH
}

module "test-sonarqube-sentry" {
  source       = "https://tfmodules.deliveroo.net/sentry_project/1.4.zip"
  project_name = local.REPO_NAME

  # Only deploy the sentry creation in the global shard
  count = var.shard == "global" ? 1 : 0
}

module "co-test-sonarqube-internal" {
  source  = "terraform-registry.deliveroo.net/deliveroo/roo-service/aws"
  version = "~> 1.0"

  service_type       = "internal_web"

  market_dns_enabled             = true
  markets_enabled_for_home_shard = [for m in data.roo_market_entries.all.items : m.market]

  app_name             = local.HOPPER_APP_NAME
  datadog_alert_target = local.DATADOG_ALERT_CHANNELS[var.env_name]

  service_name       = "example"
  container_port     = 3000
  health_check_path  = "/ping"
  health_check_codes = "200"

  subdomain_prefix = "internal-"
}

