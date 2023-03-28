module "bnt-internal-test-go-app_web" {
  source  = "terraform-registry.deliveroo.net/deliveroo/roo-service/aws"
  version = "~> 1.0"

  app_name     = module.bnt-internal-test-go-app.app_name
  service_name = "web"
  service_type = "public_web"

  datadog_alert_target = module.bnt-internal-test-go-app.datadog_alert_target

  container_port     = 3000
  health_check_path  = "/ping"
  health_check_codes = "200"
}