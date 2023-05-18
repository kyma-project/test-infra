terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "opa-gatekeeper"
  }
}
