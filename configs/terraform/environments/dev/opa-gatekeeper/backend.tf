terraform {
  required_version = ">= 1.6.1"

  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "opa-gatekeeper"
  }
}
