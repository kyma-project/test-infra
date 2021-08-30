module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getslackuserforgithubuser

go 1.14

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	cloud.google.com/go v0.91.1
	cloud.google.com/go/firestore v1.5.0
	github.com/kyma-project/test-infra v0.0.0-20210826124132-fce6cc2cee02
	k8s.io/test-infra v0.0.0-20210812232458-c6e29bb385e0 // indirect
)
