module "secrets_leaks_log_scanner" {
  depends_on = [module.slack_message_sender]

  providers = {
    google = google
  }
  source                   = "../../modules/secrets-leaks-log-scanner"
  gcp_project_id           = var.gcp_project_id
  slack_message_sender_url = module.slack_message_sender.url
}
# (2025-03-04)