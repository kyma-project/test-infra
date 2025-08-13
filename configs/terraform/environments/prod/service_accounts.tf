resource "google_service_account" "kyma-oci-image-builder" {
  account_id   = "kyma-oci-image-builder"
  display_name = "kyma-oci-image-builder"
  description  = "Service Account for retrieving secrets on the oci-image-builder ADO pipeline."
}

resource "google_service_account" "terraform-planner" {
  account_id   = "terraform-planner"
  display_name = "terraform-planner"
  description  = "Identity of terraform planner"
}

resource "google_service_account" "counduit-cli-bucket" {
  account_id   = "counduit-cli-bucket"
  display_name = "counduit-cli-bucket"
  description  = "SA to push/pull conduit-cli"
}

resource "google_service_account" "sa-prowjob-gcp-logging-client" {
  account_id   = "sa-prowjob-gcp-logging-client"
  display_name = "sa-prowjob-gcp-logging-client"
  description  = "Read, write access to google cloud logging for prowjobs. Will be removed with Prow"
}

resource "google_service_account" "terraform-executor" {
  account_id   = "terraform-executor"
  display_name = "terraform-executor"
  description  = "Identity of terraform executor."
}

resource "google_service_account" "sa-kyma-project" {
  account_id   = "sa-kyma-project"
  display_name = "sa-kyma-project"
  description  = "SA to manage PROD Artifact Registry in SAP CX Kyma Project"
}

resource "google_service_account" "gitleaks-secret-accesor" {
  account_id   = "gitleaks-secret-accesor"
  display_name = "gitleaks-secret-accesor"
  description  = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
}

resource "google_service_account" "sa-secret-update" {
  account_id   = "sa-secret-update"
  display_name = "sa-secret-update"
  description  = "Can update secrets in Secret Manager"
}

resource "google_service_account" "sa-security-dashboard-oauth" {
  account_id   = "sa-security-dashboard-oauth"
  display_name = "sa-security-dashboard-oauth"
  description  = "Used for the Security dashboard cloud run"
}

resource "google_service_account" "firebase-adminsdk-udzxq" {
  account_id   = "firebase-adminsdk-udzxq"
  display_name = "firebase-adminsdk"
  description  = "Firebase Admin SDK Service Agent"
}

resource "google_service_account" "sa-dev-kyma-project" {
  account_id   = "sa-dev-kyma-project"
  display_name = "sa-dev-kyma-project"
  description  = "SA to manage DEV Artifact Registry in SAP CX Kyma Project"
}


resource "google_service_account" "kyma-security-scanners" {
  account_id   = "kyma-security-scanners"
  display_name = "kyma-security-scanners"
  description  = "Service account for retrieving secrets on the security-scanners and orphan-cleaner Azure pipelines."
}

resource "google_service_account" "kyma-compliance-pipeline" {
  account_id   = "kyma-compliance-pipeline"
  display_name = "kyma-compliance-pipeline"
  description  = "Service account for retrieving secrets on the compliance Azure pipeline."
}

resource "google_service_account" "neighbors-conduit-cli-builder" {
  account_id   = "neighbors-conduit-cli-builder"
  display_name = "neighbors-conduit-cli-builder"
  description  = "Service account for retrieving secrets on the conduit-cli build pipeline."
}