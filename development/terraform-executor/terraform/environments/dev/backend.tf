terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "terraform-executor"
  }
}
