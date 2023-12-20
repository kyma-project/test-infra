module "github_webhook_gateway" {

  providers = {
    google = google
  }
  source         = "../../modules/github-webhook-gateway"
  gcp_project_id = var.gcp_project_id
}
