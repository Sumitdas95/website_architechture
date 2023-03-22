module "bnt-internal-test-go-app" {
  source  = "terraform-registry.deliveroo.net/deliveroo/roo_app_basic/aws"
  version = "~> 4.0"
  datadog_alert_target = {
    production = ""
    platform   = ""
    staging    = ""
    sandbox    = ""
  }

  repo_name   = "bnt-internal-test-go"
  description = "Dogfooding repo for all things BNT"
  tier        = "4"

  # Ownership data
  team_name     = "production_engineering"
  owner_email   = "" # TODO
  playbook_link = "" # TODO
  rota_link     = "" # TODO
  slack_url     = "" # TODO

  uses_feature_flags = false
  sentry_enabled     = false
}