module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getgithubcommiter

go 1.14

replace (
	k8s.io/api => k8s.io/api v0.21.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
)

require (
	cloud.google.com/go v0.91.0
	github.com/kyma-project/test-infra v0.0.0-20210826124132-fce6cc2cee02
	k8s.io/test-infra v0.0.0-20210407040951-51f95c2d525e
)
