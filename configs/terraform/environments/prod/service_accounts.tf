
resource "google_service_account" "sa-gke-kyma-integration" {
  account_id       = "sa-gke-kyma-integration"
  display_name     = "sa-gke-kyma-integration"
  description      = "Service account is used by Prow to integrate with GKE."
  disabled         = false
  email            = "sa-gke-kyma-integration@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-gke-kyma-integration@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "111305646520816790274"
  project_id       = var.gcp_project_id
  unique_id        = "111305646520816790274"
}

resource "google_service_account" "gcr-cleaner" {
  account_id       = "gcr-cleaner"
  display_name     = "gcr-cleaner"
  description      = "Service account is used by gcr-cleaner tool."
  disabled         = false
  email            = "gcr-cleaner@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/gcr-cleaner@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "109243349608620298947"
  project_id       = var.gcp_project_id
  unique_id        = "109243349608620298947"
}

resource "google_service_account" "github-issue-finder" {
  account_id       = "github-issue-finder"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running github issue finder service."
  disabled         = false
  email            = "github-issue-finder@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/github-issue-finder@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "116758462542272084643"
  project_id       = var.gcp_project_id
  unique_id        = "116758462542272084643"
}

resource "google_service_account" "secrets-leak-log-scanner" {
  account_id       = "secrets-leak-log-scanner"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running log scanner service."
  disabled         = false
  email            = "secrets-leak-log-scanner@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/secrets-leak-log-scanner@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "114460691306419275918"
  project_id       = var.gcp_project_id
  unique_id        = "114460691306419275918"
}

resource "google_service_account" "control-plane" {
  account_id       = "control-plane"
  display_name     = "control-plane"
  description      = "Main prow components SA Will be removed with Prow"
  disabled         = false
  email            = "control-plane@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/control-plane@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "109021452451580464730"
  project_id       = var.gcp_project_id
  unique_id        = "109021452451580464730"
}

resource "google_service_account" "slack-message-sender" {
  account_id       = "slack-message-sender"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running slack message sender service."
  disabled         = false
  email            = "slack-message-sender@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/slack-message-sender@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "114182019584145313167"
  project_id       = var.gcp_project_id
  unique_id        = "114182019584145313167"
}

resource "google_service_account" "kyma-oci-image-builder" {
  account_id       = "kyma-oci-image-builder"
  display_name     = "kyma-oci-image-builder"
  description      = "Service Account for retrieving secrets on the oci-image-builder ADO pipeline."
  disabled         = false
  email            = "kyma-oci-image-builder@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/kyma-oci-image-builder@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "111927626351367182700"
  project_id       = var.gcp_project_id
  unique_id        = "111927626351367182700"
}

resource "google_service_account" "sa-gardener-logs" {
  account_id       = "sa-gardener-logs"
  display_name     = "sa-gardener-logs"
  description      = "SA used by gardener cluster to send logs to Stackdriver. Will be removed with Prow"
  disabled         = false
  email            = "sa-gardener-logs@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-gardener-logs@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "115455992201052051466"
  project_id       = var.gcp_project_id
  unique_id        = "115455992201052051466"
}

resource "google_service_account" "terraform-planner" {
  account_id       = "terraform-planner"
  display_name     = "terraform-planner"
  description      = "Identity of terraform planner"
  disabled         = false
  email            = "terraform-planner@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/terraform-planner@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "112039585612554997148"
  project_id       = var.gcp_project_id
  unique_id        = "112039585612554997148"
}

resource "google_service_account" "counduit-cli-bucket" {
  account_id       = "counduit-cli-bucket"
  display_name     = "counduit-cli-bucket"
  description      = "SA to push/pull conduit-cli"
  disabled         = false
  email            = "counduit-cli-bucket@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/counduit-cli-bucket@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "111562262539619981860"
  project_id       = var.gcp_project_id
  unique_id        = "111562262539619981860"
}

resource "google_service_account" "gencred-refresher" {
  account_id       = "gencred-refresher"
  display_name     = "gencred-refresher"
  description      = "Service Account used to automatically refresh kubeconfigs for workload cluster on Prow using ProwJob `ci-gencred-refresh-kubeconfig` Will be removed with Prow"
  disabled         = false
  email            = "gencred-refresher@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/gencred-refresher@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "118113454779596169476"
  project_id       = var.gcp_project_id
  unique_id        = "118113454779596169476"
}

resource "google_service_account" "sa-prowjob-gcp-logging-client" {
  account_id       = "sa-prowjob-gcp-logging-client"
  display_name     = "sa-prowjob-gcp-logging-client"
  description      = "Read, write access to google cloud logging for prowjobs. Will be removed with Prow"
  disabled         = false
  email            = "sa-prowjob-gcp-logging-client@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-prowjob-gcp-logging-client@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "111950703097885409586"
  project_id       = var.gcp_project_id
  unique_id        = "111950703097885409586"
}

resource "google_service_account" "secret-manager-trusted" {
  account_id       = "secret-manager-trusted"
  display_name     = "secret-manager-trusted"
  description      = "Will be removed with Prow"
  disabled         = false
  email            = "secret-manager-trusted@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/secret-manager-trusted@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "102487129866240972947"
  project_id       = var.gcp_project_id
  unique_id        = "102487129866240972947"
}

resource "google_service_account" "terraform-executor" {
  account_id       = "terraform-executor"
  display_name     = "terraform-executor"
  description      = "Identity of terraform executor. It's mapped to k8s service account through workload identity."
  disabled         = false
  email            = "terraform-executor@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/terraform-executor@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "109665069699011807029"
  project_id       = var.gcp_project_id
  unique_id        = "109665069699011807029"
}

resource "google_service_account" "sa-gcr-kyma-project-trusted" {
  account_id       = "sa-gcr-kyma-project-trusted"
  display_name     = "sa-gcr-kyma-project-trusted"
  description      = "Access to GCR in kyma-project and KMS key in kyma-prow. Will be removed with Prow"
  disabled         = false
  email            = "sa-gcr-kyma-project-trusted@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-gcr-kyma-project-trusted@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "101526224201212697145"
  project_id       = var.gcp_project_id
  unique_id        = "101526224201212697145"
}

resource "google_service_account" "sa-gcs-plank" {
  account_id       = "sa-gcs-plank"
  display_name     = "sa-gcs-plank"
  description      = "The `plank` component schedules the pod requested by a prowjob. Will be removed with Prow"
  disabled         = false
  email            = "sa-gcs-plank@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-gcs-plank@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "102765494148359210250"
  project_id       = var.gcp_project_id
  unique_id        = "102765494148359210250"
}

resource "google_service_account" "sa-kyma-project" {
  account_id       = "sa-kyma-project"
  display_name     = "sa-kyma-project"
  description      = "SA to manage PROD Artifact Registry in SAP CX Kyma Project"
  disabled         = false
  email            = "sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "104982382979273080878"
  project_id       = var.gcp_project_id
  unique_id        = "104982382979273080878"
}

resource "google_service_account" "sa-prow-job-resource-cleaners" {
  account_id       = "sa-prow-job-resource-cleaners"
  display_name     = "sa-prow-job-resource-cleaners"
  description      = "SA used by multiple resource cleaner prowjobs. Will be removed with Prow"
  disabled         = false
  email            = "sa-prow-job-resource-cleaners@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-prow-job-resource-cleaners@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "102546313717271893433"
  project_id       = var.gcp_project_id
  unique_id        = "102546313717271893433"
}

resource "google_service_account" "sa-kyma-artifacts" {
  account_id       = "sa-kyma-artifacts"
  display_name     = "sa-kyma-artifacts"
  description      = "Service account used by ProwJob kyma-artifacts. Will be removed with Prow"
  disabled         = false
  email            = "sa-kyma-artifacts@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-kyma-artifacts@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "101310044122891354991"
  project_id       = var.gcp_project_id
  unique_id        = "101310044122891354991"
}

resource "google_service_account" "secrets-leak-detector" {
  account_id       = "secrets-leak-detector"
  display_name     = "default_display_name"
  description      = "Identity of secrets leak detector application."
  disabled         = false
  email            = "secrets-leak-detector@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/secrets-leak-detector@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "105851366813272921356"
  project_id       = var.gcp_project_id
  unique_id        = "105851366813272921356"
}

resource "google_service_account" "sa-keys-cleaner" {
  account_id       = "sa-keys-cleaner"
  display_name     = "default_display_name"
  description      = "Identity of the service account keys rotator service."
  disabled         = false
  email            = "sa-keys-cleaner@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-keys-cleaner@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "101317727774651823048"
  project_id       = var.gcp_project_id
  unique_id        = "101317727774651823048"
}

resource "google_service_account" "gitleaks-secret-accesor" {
  account_id       = "gitleaks-secret-accesor"
  display_name     = "gitleaks-secret-accesor"
  description      = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
  disabled         = false
  email            = "gitleaks-secret-accesor@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/gitleaks-secret-accesor@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "115976877087340028822"
  project_id       = var.gcp_project_id
  unique_id        = "115976877087340028822"
}

resource "google_service_account" "sa-secret-update" {
  account_id       = "sa-secret-update"
  display_name     = "sa-secret-update"
  description      = "Can update secrets in Secret Manager"
  disabled         = false
  email            = "sa-secret-update@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-secret-update@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "106952441418360864560"
  project_id       = var.gcp_project_id
  unique_id        = "106952441418360864560"
}

resource "google_service_account" "sa-kyma-dns-serviceuser" {
  account_id       = "sa-kyma-dns-serviceuser"
  display_name     = "sa-kyma-dns-serviceuser"
  description      = "<Used by api-gateway> Service Account used to manipulate DNS entries in sap-kyma-prow-workloads. Will be removed with Prow"
  disabled         = false
  email            = "sa-kyma-dns-serviceuser@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-kyma-dns-serviceuser@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "112558806674283527135"
  project_id       = var.gcp_project_id
  unique_id        = "112558806674283527135"
}

resource "google_service_account" "sa-security-dashboard-oauth" {
  account_id       = "sa-security-dashboard-oauth"
  display_name     = "sa-security-dashboard-oauth"
  description      = "Used for the Security dashboard cloud run"
  disabled         = false
  email            = "sa-security-dashboard-oauth@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-security-dashboard-oauth@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "100374433167316407961"
  project_id       = var.gcp_project_id
  unique_id        = "100374433167316407961"
}

resource "google_service_account" "firebase-adminsdk-udzxq" {
  account_id       = "firebase-adminsdk-udzxq"
  display_name     = "firebase-adminsdk"
  description      = "Firebase Admin SDK Service Agent"
  disabled         = false
  email            = "firebase-adminsdk-udzxq@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/firebase-adminsdk-udzxq@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "114373007268493913808"
  project_id       = var.gcp_project_id
  unique_id        = "114373007268493913808"
}

resource "google_service_account" "secret-manager-untrusted" {
  account_id       = "secret-manager-untrusted"
  display_name     = "secret-manager-untrusted"
  description      = "Will be deleted with Prow"
  disabled         = false
  email            = "secret-manager-untrusted@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/secret-manager-untrusted@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "112256620414739996658"
  project_id       = var.gcp_project_id
  unique_id        = "112256620414739996658"
}

resource "google_service_account" "sa-prow-deploy" {
  account_id       = "sa-prow-deploy"
  display_name     = "sa-prow-deploy"
  description      = "SA with admin rights in Prow cluster Will be removed with Prow"
  disabled         = false
  email            = "sa-prow-deploy@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-prow-deploy@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "114531992527279342014"
  project_id       = var.gcp_project_id
  unique_id        = "114531992527279342014"
}

resource "google_service_account" "github-issue-creator" {
  account_id       = "github-issue-creator"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running github issue creator service."
  disabled         = false
  email            = "github-issue-creator@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/github-issue-creator@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "116199434171095445870"
  project_id       = var.gcp_project_id
  unique_id        = "116199434171095445870"
}

resource "google_service_account" "sa-dev-kyma-project" {
  account_id       = "sa-dev-kyma-project"
  display_name     = "sa-dev-kyma-project"
  description      = "SA to manage DEV Artifact Registry in SAP CX Kyma Project"
  disabled         = false
  email            = "sa-dev-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-dev-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "104900265353366440226"
  project_id       = var.gcp_project_id
  unique_id        = "104900265353366440226"
}

resource "google_service_account" "github-webhook-gateway" {
  account_id       = "github-webhook-gateway"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running github webhook gateway service."
  disabled         = false
  email            = "github-webhook-gateway@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/github-webhook-gateway@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "108309727809268268116"
  project_id       = var.gcp_project_id
  unique_id        = "108309727809268268116"
}

resource "google_service_account" "sa-keys-rotator" {
  account_id       = "sa-keys-rotator"
  display_name     = "default_display_name"
  description      = "Identity of the service account keys rotator service."
  disabled         = false
  email            = "sa-keys-rotator@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-keys-rotator@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "116267434130697196528"
  project_id       = var.gcp_project_id
  unique_id        = "116267434130697196528"
}

resource "google_service_account" "gcs-bucket-mover" {
  account_id       = "gcs-bucket-mover"
  display_name     = "default_display_name"
  description      = "Identity of cloud run instance running gcs bucket mover service."
  disabled         = false
  email            = "gcs-bucket-mover@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/gcs-bucket-mover@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "109869306046945738408"
  project_id       = var.gcp_project_id
  unique_id        = "109869306046945738408"
}

resource "google_service_account" "secret-manager-prow" {
  account_id       = "secret-manager-prow"
  display_name     = "secret-manager-prow"
  description      = "Will be removed with Prow"
  disabled         = false
  email            = "secret-manager-prow@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/secret-manager-prow@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "110693330049686538177"
  project_id       = var.gcp_project_id
  unique_id        = "110693330049686538177"
}

resource "google_service_account" "sa-vm-kyma-integration" {
  account_id       = "sa-vm-kyma-integration"
  display_name     = "sa-vm-kyma-integration"
  description      = "Will be removed with Prow"
  disabled         = false
  email            = "sa-vm-kyma-integration@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-vm-kyma-integration@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "117798148653314453801"
  project_id       = var.gcp_project_id
  unique_id        = "117798148653314453801"
}

resource "google_service_account" "sa-prow-pubsub" {
  account_id       = "sa-prow-pubsub"
  display_name     = "sa-prow-pubsub"
  description      = "Run prow related pubsub topics, subscriptions and cloud functions. Will be removed with Prow"
  disabled         = false
  email            = "sa-prow-pubsub@sap-kyma-prow.iam.gserviceaccount.com"
  etag             = "MDEwMjE5MjA="
  name             = "projects/sap-kyma-prow/serviceAccounts/sa-prow-pubsub@sap-kyma-prow.iam.gserviceaccount.com"
  oauth2_client_id = "100524274165017892082"
  project_id       = var.gcp_project_id
  unique_id        = "100524274165017892082"
}
