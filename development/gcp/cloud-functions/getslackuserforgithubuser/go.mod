module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getslackuserforgithubuser

go 1.14

replace github.com/kyma-project/test-infra => ../../../..
require (
	cloud.google.com/go v0.91.1
	github.com/kyma-project/test-infra v0.0.0-20210816141126-69d574c7dec6 // indirect
	k8s.io/test-infra v0.0.0-20210812232458-c6e29bb385e0
)
