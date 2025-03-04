resource "google_dns_managed_zone" "build_kyma" {
  dns_name = "build.kyma-project.io."
  name     = "build-kyma"
  dnssec_config {
    state = "on"
  }
}
# (2025-03-04)