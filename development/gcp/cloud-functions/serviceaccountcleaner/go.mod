module github.com/kyma-project/test-infra/development/gcp/cloud-functions/serviceaccountcleaner

go 1.16

replace (
	github.com/kyma-project/test-infra v0.0.0-20220715122928-d02a288f4078 => ../../../../.
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	cloud.google.com/go/compute v1.7.0
	github.com/GoogleCloudPlatform/functions-framework-go v1.6.1
	github.com/kyma-project/test-infra v0.0.0-20220715122928-d02a288f4078
	google.golang.org/api v0.94.0
)
