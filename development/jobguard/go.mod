module github.com/kyma-project/test-infra/development/jobguard

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go => k8s.io/client-go v0.16.4
)

require (
	github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/test-infra v0.0.0-20200331085241-bf7dc9346358
)
