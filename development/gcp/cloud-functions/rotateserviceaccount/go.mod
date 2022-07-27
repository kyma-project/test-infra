module github.com/kyma-project/test-infra/development/gcp/cloud-functions/rotateserviceaccount

go 1.18

require (
	github.com/kyma-project/test-infra v0.0.0-20220715122928-d02a288f4078
)

replace (
    github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager v0.0.0-20220715122928-d02a288f4078 => github.com/Halamix2/test-infra/development/gcp/pkg/secretmanager cloudfunc_secret_rotate
    github.com/kyma-project/test-infra/development/gcp/pkg/secretversionsmanager v0.0.0-20220715122928-d02a288f4078 => github.com/Halamix2/test-infra/development/gcp/pkg/secretversionsmanager cloudfunc_secret_rotate
)
