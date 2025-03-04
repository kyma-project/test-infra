module "cors_proxy" {

  providers = {
    google = google
  }
  source         = "../../modules/cors-proxy"
  gcp_project_id = var.gcp_project_id
}
# (2025-03-04)