module "cors_proxy" {

  providers = {
    google = google
  }
  source         = "../../modules/cors-proxy"
}
