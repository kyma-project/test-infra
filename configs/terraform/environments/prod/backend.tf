terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "kyma-test-infra-prod"
  }
}
# (2025-03-04)