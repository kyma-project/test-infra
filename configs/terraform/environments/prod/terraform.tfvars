kyma_project_artifact_registry_collection = {
  modules-internal = {
    name = "modules-internal"
    owner = "neighbors"
    type = "production"
    description = "modules-internal registry"
    reader_serviceaccounts = [
      "klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"
    ]
    repoAdmin_serviceaccounts = ["kyma-submission-pipeline@kyma-project.iam.gserviceaccount.com"]
  },
}
service_account_keys_rotator_service_name            = "service-account-keys-rotator"
service_account_keys_rotator_image                   = "europe-docker.pkg.dev/kyma-project/prod/test-infra/rotate-service-account:v20250505-e829be8a" #gitleaks:allow
service_account_keys_cleaner_service_name            = "service-account-keys-cleaner"
service_account_keys_cleaner_image                   = "europe-docker.pkg.dev/kyma-project/prod/test-infra/service-account-keys-cleaner:v20250505-e829be8a" #gitleaks:allow
service_account_key_latest_version_min_age           = 24
service_account_keys_cleaner_scheduler_cron_schedule = "0 0 * * 1-5"

signify_secret_rotator_service_name = "signify-secret-rotator"
signify_secret_rotator_image        = "europe-docker.pkg.dev/kyma-project/prod/test-infra/signify-secret-rotator:v20250422-e7b3876b" #gitleaks:allow
