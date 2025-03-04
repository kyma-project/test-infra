terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "secrets-rotator"
  }
}
# (2025-03-04)