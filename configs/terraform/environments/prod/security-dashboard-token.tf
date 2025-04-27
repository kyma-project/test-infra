module "security_dashboard_token" {

  providers = {
    google = google
  }
  source         = "../../modules/security-dashboard-token"
}
