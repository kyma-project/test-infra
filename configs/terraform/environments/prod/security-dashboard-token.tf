module "cors_proxy" {

  providers = {
    google = google
  }
  source         = "../../modules/security-dashboard-token"
  gcp_project_id = var.gcp_project_id
}
