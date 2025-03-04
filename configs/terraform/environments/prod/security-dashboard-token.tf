module "security_dashboard_token" {

  providers = {
    google = google
  }
  source         = "../../modules/security-dashboard-token"
  gcp_project_id = var.gcp_project_id
}
# (2025-03-04)