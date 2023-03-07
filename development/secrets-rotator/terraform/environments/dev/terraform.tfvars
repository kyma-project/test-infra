project_id                                 = "sap-kyma-neighbors-dev"
location                                   = "europe-west3"
service_account_keys_rotator_service_name  = "service-account-keys-rotator"
service_account_keys_rotator_image         = "europe-docker.pkg.dev/kyma-project/prod/test-infra/rotate-service-account:v20230307-cf164cd1"
service_account_keys_cleaner_service_name  = "service-account-keys-cleaner"
service_account_keys_cleaner_image         = "europe-docker.pkg.dev/kyma-project/prod/test-infra/service-account-keys-cleaner:v20230301-6267d66d"
service_account_key_latest_version_min_age = 24
