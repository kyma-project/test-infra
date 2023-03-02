project_id                                = "sap-kyma-neighbors-dev"
location                                  = "europe-west3"
service_name                              = "service-account-keys-rotator"
create_dead_letter_topic                  = true
create_secret_manager_notifications_topic = true
service_account_keys_rotator_image        = "europe-docker.pkg.dev/kyma-project/prod/test-infra/rotate-service-account:v20230301-6267d66d"
