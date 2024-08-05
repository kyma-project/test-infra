resource "google_service_account" "sa-gke-kyma-integration" {
  account_id   = "sa-gke-kyma-integration"
  display_name = "sa-gke-kyma-integration"
  description  = "Service account is used by Prow to integrate with GKE. Will be removed with Prow"
}

resource "google_service_account" "gcr-cleaner" {
  account_id   = "gcr-cleaner"
  display_name = "gcr-cleaner"
  description  = "Service account is used by gcr-cleaner tool."
}

resource "google_service_account" "control-plane" {
  account_id   = "control-plane"
  display_name = "control-plane"
  description  = "Main prow components SA Will be removed with Prow"
}

resource "google_service_account" "kyma-oci-image-builder" {
  account_id   = "kyma-oci-image-builder"
  display_name = "kyma-oci-image-builder"
  description  = "Service Account for retrieving secrets on the oci-image-builder ADO pipeline."
}

resource "google_service_account" "sa-gardener-logs" {
  account_id   = "sa-gardener-logs"
  display_name = "sa-gardener-logs"
  description  = "SA used by gardener cluster to send logs to Stackdriver. Will be removed with Prow"
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

resource "google_service_account" "gencred-refresher" {
  account_id   = "gencred-refresher"
  display_name = "gencred-refresher"
  description  = "Service Account used to automatically refresh kubeconfigs for workload cluster on Prow using ProwJob `ci-gencred-refresh-kubeconfig` Will be removed with Prow"
}

resource "google_service_account" "sa-prowjob-gcp-logging-client" {
  account_id   = "sa-prowjob-gcp-logging-client"
  display_name = "sa-prowjob-gcp-logging-client"
  description  = "Read, write access to google cloud logging for prowjobs. Will be removed with Prow"
}

resource "google_service_account" "secret-manager-trusted" {
  account_id   = "secret-manager-trusted"
  display_name = "secret-manager-trusted"
  description  = "Will be removed with Prow"
}

resource "google_service_account" "terraform-executor" {
  account_id   = "terraform-executor"
  display_name = "terraform-executor"
  description  = "Identity of terraform executor. It's mapped to k8s service account through workload identity."
}

resource "google_service_account" "sa-gcr-kyma-project-trusted" {
  account_id   = "sa-gcr-kyma-project-trusted"
  display_name = "sa-gcr-kyma-project-trusted"
  description  = "Access to GCR in kyma-project and KMS key in kyma-prow. Will be removed with Prow"
}

resource "google_service_account" "sa-gcs-plank" {
  account_id   = "sa-gcs-plank"
  display_name = "sa-gcs-plank"
  description  = "The `plank` component schedules the pod requested by a prowjob. Will be removed with Prow"
}

resource "google_service_account" "sa-kyma-project" {
  account_id   = "sa-kyma-project"
  display_name = "sa-kyma-project"
  description  = "SA to manage PROD Artifact Registry in SAP CX Kyma Project"
}

resource "google_service_account" "sa-prow-job-resource-cleaners" {
  account_id   = "sa-prow-job-resource-cleaners"
  display_name = "sa-prow-job-resource-cleaners"
  description  = "SA used by multiple resource cleaner prowjobs. Will be removed with Prow"
}

resource "google_service_account" "sa-kyma-artifacts" {
  account_id   = "sa-kyma-artifacts"
  display_name = "sa-kyma-artifacts"
  description  = "Service account used by ProwJob kyma-artifacts. Will be removed with Prow"
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

resource "google_service_account" "sa-kyma-dns-serviceuser" {
  account_id   = "sa-kyma-dns-serviceuser"
  display_name = "sa-kyma-dns-serviceuser"
  description  = "<Used by api-gateway> Service Account used to manipulate DNS entries in sap-kyma-prow-workloads. Will be removed with Prow"
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

resource "google_service_account" "secret-manager-untrusted" {
  account_id   = "secret-manager-untrusted"
  display_name = "secret-manager-untrusted"
  description  = "Will be deleted with Prow"
}

resource "google_service_account" "sa-prow-deploy" {
  account_id   = "sa-prow-deploy"
  display_name = "sa-prow-deploy"
  description  = "SA with admin rights in Prow cluster Will be removed with Prow"
}

resource "google_service_account" "sa-dev-kyma-project" {
  account_id   = "sa-dev-kyma-project"
  display_name = "sa-dev-kyma-project"
  description  = "SA to manage DEV Artifact Registry in SAP CX Kyma Project"
}

resource "google_service_account" "secret-manager-prow" {
  account_id   = "secret-manager-prow"
  display_name = "secret-manager-prow"
  description  = "Will be removed with Prow"
}

resource "google_service_account" "sa-vm-kyma-integration" {
  account_id   = "sa-vm-kyma-integration"
  display_name = "sa-vm-kyma-integration"
  description  = "Will be removed with Prow"
}

resource "google_service_account" "sa-prow-pubsub" {
  account_id   = "sa-prow-pubsub"
  display_name = "sa-prow-pubsub"
  description  = "Run prow related pubsub topics, subscriptions and cloud functions. Will be removed with Prow"
}

resource "google_service_account" "kyma-submission-pipeline" {
  account_id   = "kyma-submission-pipeline"
  display_name = "kyma-submission-pipeline"
  description  = "Service Account for retrieving secrets on the submission-pipeline ADO pipeline."
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
