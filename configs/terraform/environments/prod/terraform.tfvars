service_account_keys_rotator_service_name            = "service-account-keys-rotator"
service_account_keys_rotator_image                   = "europe-docker.pkg.dev/kyma-project/prod/test-infra/rotate-service-account:v20251014-529e8502" #gitleaks:allow
service_account_keys_cleaner_service_name            = "service-account-keys-cleaner"
service_account_keys_cleaner_image                   = "europe-docker.pkg.dev/kyma-project/prod/test-infra/service-account-keys-cleaner:v20251016-c243dd68" #gitleaks:allow
service_account_key_latest_version_min_age           = 24
service_account_keys_cleaner_scheduler_cron_schedule = "0 0 * * 1-5"

signify_secret_rotator_service_name = "signify-secret-rotator"
signify_secret_rotator_image        = "europe-docker.pkg.dev/kyma-project/prod/test-infra/signify-secret-rotator:v20251014-c3133fef" #gitleaks:allow
