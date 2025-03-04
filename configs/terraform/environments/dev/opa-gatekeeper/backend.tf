terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "opa-gatekeeper"
  }
}
# (2025-03-04)