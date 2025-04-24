terraform {
  required_version = ">= 1.8.0"

  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "opa-gatekeeper"
  }
}
