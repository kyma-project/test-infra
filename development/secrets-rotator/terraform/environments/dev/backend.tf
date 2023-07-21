terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "secrets-rotator"
  }
}
